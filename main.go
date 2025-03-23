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
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"unsafe"
)

var (
	difficulty uint
	listenAddr string
	destHost   string
	secondary  bool
)

func init() {
	flag.UintVar(&difficulty, "difficulty", 17, "leading zero bits required for the challenge")
	flag.StringVar(&listenAddr, "listen", ":8081", "address to listen on")
	flag.StringVar(&destHost, "upstream", "http://127.0.0.1:8080", "destination url base to proxy to")
	flag.BoolVar(&secondary, "secondary", false, "trust X-Forwarded-For headers")
	flag.Parse()
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

var reverseProxy *httputil.ReverseProxy

func init() {
	parsedURL, err := url.Parse(destHost)
	if err != nil {
		log.Fatal(err)
	}
	reverseProxy = httputil.NewSingleHostReverseProxy(parsedURL)
}

type tparams struct {
	UnsignedTokenBase64 string
	NeedBits            uint
	Message             string
}

func main() {
	log.Fatal(http.ListenAndServe(listenAddr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println(getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))

		cookie, err := request.Cookie("powxy")
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				http.Error(writer, "error fetching cookie", http.StatusInternalServerError)
				return
			}
		}

		expectedToken := makeSignedToken(request)

		if validateCookie(cookie, expectedToken) {
			proxyRequest(writer, request)
			return
		}

		authPage := func(message string) {
			_ = tmpl.Execute(writer, tparams{
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

		if len(nonce) > 32 {
			authPage("Your submission was too long.")
			return
		}

		h := sha256.New()
		h.Write(expectedToken[:sha256.Size])
		h.Write(nonce)
		ck := h.Sum(nil)
		if !validateBitZeros(ck, difficulty) {
			authPage("Your submission was incorrect, or your session has expired while submitting.")
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

func getRemoteIP(request *http.Request) (remoteIP string) {
	if secondary {
		remoteIP, _, _ = strings.Cut(request.Header.Get("X-Forwarded-For"), ",")
	}
	if remoteIP == "" {
		remoteIP = request.RemoteAddr
		index := strings.LastIndex(remoteIP, ":")
		if index != -1 {
			remoteIP = remoteIP[:index]
		}
	}
	return
}

func makeSignedToken(request *http.Request) []byte {
	buf := make([]byte, 0, 2*sha256.Size)

	timeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(timeBuf, time.Now().Unix()/604800)

	remoteIP := getRemoteIP(request)

	h := sha256.New()
	h.Write(timeBuf)
	h.Write(stringToBytes(remoteIP))
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
	reverseProxy.ServeHTTP(writer, request)
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
