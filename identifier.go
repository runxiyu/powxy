// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"time"
)

// makeIdentifierMAC generates an identifier that semi-uniquely identifies the client,
// and generates a MAC for that identifier.
func makeIdentifierMAC(request *http.Request) (identifier []byte, mac []byte) {
	identifier = make([]byte, 0, sha256.Size)
	mac = make([]byte, 0, sha256.Size)

	timeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(timeBuf, time.Now().Unix()/604800)

	remoteIP := getRemoteIP(request)

	// It is safe to use stringToBytes here as h.Write does not modify its
	// argument.
	h := sha256.New()
	h.Write(timeBuf)
	h.Write(stringToBytes(remoteIP))
	h.Write(stringToBytes(request.Header.Get("User-Agent")))
	h.Write(stringToBytes(request.Header.Get("Accept-Encoding")))
	h.Write(stringToBytes(request.Header.Get("Accept-Language")))
	h.Write(privkeyHash)
	identifier = h.Sum(identifier)

	m := hmac.New(sha256.New, privkey)
	m.Write(identifier)
	mac = m.Sum(mac)

	return
}

// validateCookie checks if the cookie is valid by comparing the base64-decoded
// value of the cookie with an expected MAC.
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
