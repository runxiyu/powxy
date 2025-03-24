// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"log"
)

var (
	// The private key used to HMAC the challenge.
	privkey = make([]byte, 32)

	// The hash of the private key. We use this as an element of the
	// identifier.
	privkeyHash = make([]byte, 0, sha256.Size)
)

// This init generates a random private key and its hash.
func init() {
	if _, err := rand.Read(privkey); err != nil {
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write(privkey)
	privkeyHash = h.Sum(nil)
}
