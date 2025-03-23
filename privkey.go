// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"log"
)

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
