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
	flag.UintVar(&difficulty, "difficulty", 17, "leading zero bits required for the challenge")
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
<summary>Program to solve this</summary>
<pre>` + html.EscapeString(solverProgram) + `</pre>
</details>
<p id="solver_status"></p>
</body>
<script>
document.addEventListener("DOMContentLoaded", function() {
	let challenge_b64 = "{{.UnsignedTokenBase64}}";
	let difficulty = {{.NeedBits}};
	let form = document.querySelector("form");
	let field = form.querySelector("input[name='powxy']");
	let status_el = document.getElementById("solver_status");

	if (!window.crypto || !window.crypto.subtle) {
		status_el.textContent = "SubtleCrypto not available. You must solve the challenge externally.";
		return;
	}
	status_el.textContent = "SubtleCrypto detected. Attempting to solve in JS...";

	let solver_active = true;
	form.addEventListener("submit", function() {
		solver_active = false;
	});

	async function solve_pow() {
		let token_bytes = Uint8Array.from(
			atob(challenge_b64),
			ch => ch.charCodeAt(0)
		);

		let nonce = 0n;
		let buf = new ArrayBuffer(8);
		let view = new DataView(buf);

		while (solver_active) {
			view.setBigUint64(0, nonce, true);

			let candidate = new Uint8Array(token_bytes.length + 8);
			candidate.set(token_bytes, 0);
			candidate.set(new Uint8Array(buf), token_bytes.length);

			let digest_buffer = await crypto.subtle.digest("SHA-256", candidate);
			let digest = new Uint8Array(digest_buffer);

			if (has_leading_zero_bits(digest, difficulty)) {
				let nonce_str = String.fromCharCode(...new Uint8Array(buf));
				field.value = btoa(nonce_str);

				status_el.textContent = "Solution found. Submitting...";
				form.submit();
				return;
			}

			nonce++;

			// Update status every 256 tries
			if ((nonce & 0x00FFn) === 0n) {
				status_el.textContent = "Tried " + nonce + " candidates so far...";
				await new Promise(r => setTimeout(r, 0));
			}
		}
	}

	function has_leading_zero_bits(digest, bits) {
		let full_bytes = bits >>> 3;
		for (let i = 0; i < full_bytes; i++) {
			if (digest[i] !== 0) {
				return false;
			}
		}
		let remainder = bits & 7;
		if (remainder !== 0) {
			let mask = 0xFF << (8 - remainder);
			if ((digest[full_bytes] & mask) !== 0) {
				return false;
			}
		}
		return true;
	}

	solve_pow();
});
</script>
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
	_, _ = io.Copy(writer, response.Body)
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

const solverProgram = `
#include <openssl/evp.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <stdlib.h>
#include <errno.h>

bool validate_bit_zeros(const unsigned char *bs, uint8_t n)
{
	uint8_t q = n / 8;
	uint8_t r = n % 8;

	for (uint8_t i = 0; i < q; i++) {
		if (bs[i] != 0)
			return false;
	}

	if (r > 0) {
		unsigned char mask = (unsigned char)(0xFF << (8 - r));
		if (bs[q] & mask)
			return false;
	}

	return true;
}

int main(int argc, char **argv)
{
	if (argc < 3) {
		fprintf(stderr, "usage: %s <base64_data> <difficulty>\n",
			argv[0]);
		return 1;
	}

	size_t base64_data_len = strlen(argv[1]);
	unsigned char *base64_data = malloc(base64_data_len);
	if (!base64_data) {
		perror("malloc");
		return 1;
	}
	memcpy(base64_data, argv[1], base64_data_len);

	char *endptr = NULL;
	errno = 0;
	unsigned long tmp_val = strtoul(argv[2], &endptr, 10);
	if ((errno == ERANGE && tmp_val == ULONG_MAX) || *endptr != '\0'
	    || tmp_val > UINT8_MAX) {
		fprintf(stderr, "invalid difficulty value\n");
		free(base64_data);
		return 1;
	}
	uint8_t difficulty = (uint8_t) tmp_val;

	BIO *b64 = BIO_new(BIO_f_base64());
	BIO *bmem = BIO_new_mem_buf(base64_data, (int)base64_data_len);
	if (!b64 || !bmem) {
		fprintf(stderr, "BIO_new/BIO_new_mem_buf\n");
		free(base64_data);
		return 1;
	}

	BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
	b64 = BIO_push(b64, bmem);

	size_t decoded_cap = base64_data_len;
	unsigned char *decoded = malloc(decoded_cap);
	if (!decoded) {
		perror("malloc");
		BIO_free_all(b64);
		free(base64_data);
		return 1;
	}

	int decoded_len = BIO_read(b64, decoded, (int)decoded_cap);
	if (decoded_len < 0) {
		fprintf(stderr, "BIO_read\n");
		BIO_free_all(b64);
		free(base64_data);
		free(decoded);
		return 1;
	}
	BIO_free_all(b64);
	free(base64_data);

	EVP_MD_CTX *mdctx = EVP_MD_CTX_new();
	if (!mdctx) {
		fprintf(stderr, "EVP_MD_CTX_new\n");
		free(decoded);
		return 1;
	}

	size_t len = EVP_MD_size(EVP_sha256());
	unsigned char digest[EVP_MAX_MD_SIZE];
	size_t next = 0;

	while (1) {
		if (EVP_DigestInit_ex(mdctx, EVP_sha256(), NULL) != 1) {
			fprintf(stderr, "EVP_DigestInit_ex\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestUpdate(mdctx, decoded, decoded_len) != 1) {
			fprintf(stderr, "EVP_DigestUpdate(data)\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestUpdate(mdctx, &next, sizeof(next)) != 1) {
			fprintf(stderr, "EVP_DigestUpdate(next)\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestFinal_ex(mdctx, digest, NULL) != 1) {
			fprintf(stderr, "EVP_DigestFinal_ex\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (validate_bit_zeros(digest, difficulty)) {
			break;
		}
		next++;
		if (!next) {
			fprintf(stderr, "unsigned integer overflow\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
	}
	EVP_MD_CTX_free(mdctx);
	free(decoded);

	BIO *b64_out = BIO_new(BIO_f_base64());
	BIO *bmem_out = BIO_new(BIO_s_mem());
	if (!b64_out || !bmem_out) {
		fprintf(stderr, "BIO_new\n");
		if (b64_out)
			BIO_free_all(b64_out);
		if (bmem_out)
			BIO_free(bmem_out);
		return 1;
	}
	BIO_set_flags(b64_out, BIO_FLAGS_BASE64_NO_NL);
	b64_out = BIO_push(b64_out, bmem_out);

	if (BIO_write(b64_out, &next, sizeof(next)) < 0) {
		fprintf(stderr, "BIO_write\n");
		BIO_free_all(b64_out);
		return 1;
	}
	if (BIO_flush(b64_out) < 1) {
		fprintf(stderr, "BIO_flush\n");
		BIO_free_all(b64_out);
		return 1;
	}

	BUF_MEM *bptr = NULL;
	BIO_get_mem_ptr(b64_out, &bptr);
	if (!bptr || !bptr->data) {
		fprintf(stderr, "BIO_get_mem_ptr\n");
		BIO_free_all(b64_out);
		return 1;
	}

	write(STDOUT_FILENO, bptr->data, bptr->length);
	write(STDOUT_FILENO, "\n", 1);

	BIO_free_all(b64_out);
	return 0;
}
`
