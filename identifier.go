// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"net/http"
	"time"
)

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
