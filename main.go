// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
)

type tparams struct {
	Identifier string
	Message    string
	Global     any
}

func main() {
	log.Fatal(http.ListenAndServe(listenAddr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie("powxy")
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				log.Println("COOKIE_ERR", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
				http.Error(writer, "error fetching cookie", http.StatusInternalServerError)
				return
			}
		}

		identifier, expectedMAC := makeIdentifierMAC(request)

		if validateCookie(cookie, expectedMAC) {
			log.Println("PROXY", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			proxyRequest(writer, request)
			return
		}

		authPage := func(message string) {
			err := tmpl.Execute(writer, tparams{
				Identifier: base64.StdEncoding.EncodeToString(identifier),
				Message:    message,
				Global:     global,
			})
			if err != nil {
				log.Println("Error executing template:", err)
			}
		}

		if request.ParseForm() != nil {
			log.Println("MALFORMED", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("You submitted a malformed form.")
			return
		}

		formValues, ok := request.PostForm["powxy"]
		if !ok {
			log.Println("CHALLENGE", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("")
			return
		} else if len(formValues) != 1 {
			log.Println("FORM_VALUES", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("You submitted an invalid number of form values.")
			return
		}

		nonce, err := base64.StdEncoding.DecodeString(formValues[0])
		if err != nil {
			log.Println("BASE64", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("Your submission was improperly encoded.")
			return
		}

		if len(nonce) > 32 {
			log.Println("TOO_LONG", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("Your submission was too long.")
			return
		}

		h := sha256.New()
		h.Write(identifier)
		h.Write(nonce)
		ck := h.Sum(nil)
		if !validateBitZeros(ck, global.NeedBits) {
			log.Println("WRONG", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
			authPage("Your submission was incorrect, or your session has expired while submitting.")
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:   "powxy",
			Value:  base64.StdEncoding.EncodeToString(expectedMAC),
			Secure: true,
		})

		log.Println("ACCEPTED", getRemoteIP(request), request.RequestURI, request.Header.Get("User-Agent"))
		http.Redirect(writer, request, "", http.StatusSeeOther)
	})))
}

func validateCookie(cookie *http.Cookie, expectedMAC []byte) bool {
	if cookie == nil {
		return false
	}

	gotMAC, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(gotMAC, expectedMAC) == 1
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
