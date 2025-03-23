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
