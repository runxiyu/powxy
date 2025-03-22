package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"html"
	"html/template"
	"io"
	"log"
	"maps"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

var (
	difficulty uint
	listenAddr string
	destHost   string
)

func init() {
	flag.UintVar(&difficulty, "difficulty", 20, "leading zero bits required for the challenge")
	flag.StringVar(&listenAddr, "listen", ":8081", "address to listen on")
	flag.StringVar(&destHost, "host", "127.0.0.1:8080", "destination host to proxy to")
	flag.Parse()
}

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
}

var (
	privkey     = make([]byte, 32)
	privkeyHash = make([]byte, 0, sha256.Size)
)

func init() {
	if _, err := rand.Read(privkey); err != nil {
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write(privkey)
	privkeyHash = h.Sum(nil)
}

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("powxy").Parse(`
<!DOCTYPE html>
<html>
<head>
<title>Proof of Work Challenge</title>
</head>
<body>
<h1>Proof of Work Challenge</h1>
<p>This site is protected by <a href="https://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/">Powxy</a>.</p>
<p>You must complete this proof of work challenge before you could access this site.</p>
{{- if .Message }}
<p><strong>{{ .Message }}</strong></p>
{{- end }}
<p>Select a value, such that when it is appended to the decoded form of the following base64 string, and a SHA-256 hash is taken as a whole, the first {{ .NeedBits }} bits of the SHA-256 hash are zeros. Within one octet, higher bits are considered to be in front of lower bits.</p>
<p>{{ .UnsignedTokenBase64 }}</p>
<form method="POST">
<p>
Encode your selected value in base64 and submit it below:
</p>
<input name="powxy" type="text" />
<input type="submit" value="Submit" />
</form>
<br />
<details>
<summary>
Python script to automatically solve this
</summary><pre>` + html.EscapeString(pythonScript) + `</pre>
</details>
</body>
</html>
`)
	if err != nil {
		log.Fatal(err)
	}
}

type tparams struct {
	UnsignedTokenBase64 string
	NeedBits            uint
	Message             string
}

func main() {
	log.Fatal(http.ListenAndServe(listenAddr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.RemoteAddr, request.RequestURI)

		cookie, err := request.Cookie("powxy")
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				http.Error(writer, "error fetching cookie", http.StatusInternalServerError)
			}
		}

		expectedToken := makeSignedToken(request)

		if validateCookie(cookie, expectedToken) {
			proxyRequest(writer, request)
			return
		}

		authPage := func(message string) {
			tmpl.Execute(writer, tparams{
				UnsignedTokenBase64: base64.StdEncoding.EncodeToString(expectedToken[:sha256.Size]),
				Message:             message,
				NeedBits:            difficulty,
			})
		}

		if request.ParseForm() != nil {
			authPage("You submitted a malformed form.")
			return
		}

		formValues, ok := request.PostForm["powxy"]
		if !ok {
			authPage("")
			return
		} else if len(formValues) != 1 {
			authPage("You submitted an invalid number of form values.")
			return
		}

		nonce, err := base64.StdEncoding.DecodeString(formValues[0])
		if err != nil {
			authPage("Your submission was improperly encoded.")
			return
		}

		h := sha256.New()
		h.Write(expectedToken[:sha256.Size])
		h.Write(nonce)
		ck := h.Sum(nil)
		if !validateBitZeros(ck, difficulty) {
			authPage("Your submission was incorrect.")
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:  "powxy",
			Value: base64.StdEncoding.EncodeToString(expectedToken),
		})

		http.Redirect(writer, request, "", http.StatusSeeOther)
	})))
}

func validateCookie(cookie *http.Cookie, expectedToken []byte) bool {
	if cookie == nil {
		return false
	}

	gotToken, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(gotToken, expectedToken) == 1
}
func makeSignedToken(request *http.Request) []byte {
	buf := make([]byte, 0, 2*sha256.Size)

	timeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(timeBuf, time.Now().Unix()/604800)

	remoteAddr, _, _ := strings.Cut(request.RemoteAddr, ":")

	h := sha256.New()
	h.Write(timeBuf)
	h.Write(stringToBytes(remoteAddr))
	h.Write(stringToBytes(request.Header.Get("User-Agent")))
	h.Write(stringToBytes(request.Header.Get("Accept-Encoding")))
	h.Write(stringToBytes(request.Header.Get("Accept-Language")))
	h.Write(privkeyHash)
	buf = h.Sum(buf)
	if len(buf) != sha256.Size {
		panic("unexpected buffer length after hashing contents")
	}

	mac := hmac.New(sha256.New, privkey)
	mac.Write(buf)
	buf = mac.Sum(buf)
	if len(buf) != 2*sha256.Size {
		panic("unexpected buffer length after hmac")
	}

	return buf
}

func proxyRequest(writer http.ResponseWriter, request *http.Request) {
	request.Host = destHost
	request.URL.Host = destHost
	request.URL.Scheme = "http"
	request.RequestURI = ""

	response, err := client.Do(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		return
	}

	maps.Copy(writer.Header(), response.Header)
	writer.WriteHeader(response.StatusCode)
	io.Copy(writer, response.Body)
}

func stringToBytes(s string) (bytes []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func validateBitZeros(bs []byte, n uint) bool {
	q := n / 8
	r := n % 8

	for i := uint(0); i < q; i++ {
		if bs[i] != 0 {
			return false
		}
	}

	if r > 0 {
		mask := byte(0xFF << (8 - r))
		if bs[q]&mask != 0 {
			return false
		}
	}

	return true
}

const pythonScript = `
import base64
import hashlib
import random
import sys

def proc(encoded, n_bits):
    decoded = base64.b64decode(encoded)
    while True:
        appended = bytes([random.randint(0, 255) for _ in range(16)])
        sha256_hash = hashlib.sha256(decoded + appended).digest()
        hash_bits = ''.join(f'{byte:08b}' for byte in sha256_hash)
        if hash_bits[:n_bits] == '0' * n_bits:
            return appended

def main() :
    if len(sys.argv) != 3:
        print(f"usage: {sys.argv[0]} <encoded> <n_bits>", file=sys.stderr)
        sys.exit(1)
    
    encoded = sys.argv[1]
    n_bits = int(sys.argv[2])
    
    if n_bits < 0 or n_bits > 256:
        print("error: n_bits should be between 0 and 256", file=sys.stderr)
        sys.exit(1)
    
    print(base64.b64encode(proc(encoded, n_bits)).decode("ascii"))

if __name__ == "__main__":
    main()
`
