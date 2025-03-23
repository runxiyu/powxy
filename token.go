// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"net/http"
	"time"
)

func makeSignedToken(request *http.Request) (identifier []byte, mac []byte) {
	identifier = make([]byte, 0, sha256.Size)
	mac = make([]byte, 0, sha256.Size)

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
	identifier = h.Sum(identifier)
	if len(identifier) != sha256.Size {
		panic("unexpected buffer length after hashing contents")
	}

	m := hmac.New(sha256.New, privkey)
	m.Write(identifier)
	mac = m.Sum(mac)
	if len(mac) != sha256.Size {
		panic("unexpected buffer length after hmac")
	}

	return
}
