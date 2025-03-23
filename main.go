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
	Message                  string
	Global                   any
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

		identifier, expectedMAC := makeIdentifierMAC(request)

		if validateCookie(cookie, expectedMAC) {
			proxyRequest(writer, request)
			return
		}

		authPage := func(message string) {
			_ = tmpl.Execute(writer, tparams{
				Identifier: base64.StdEncoding.EncodeToString(identifier),
				Message:                  message,
				Global:                   global,
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
		h.Write(identifier)
		h.Write(nonce)
		ck := h.Sum(nil)
		if !validateBitZeros(ck, global.NeedBits) {
			authPage("Your submission was incorrect, or your session has expired while submitting.")
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:  "powxy",
			Value: base64.StdEncoding.EncodeToString(expectedMAC),
		})

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
