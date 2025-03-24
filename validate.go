// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/sha256"
)

// validateNonce checks if the nonce for the proof of work challenge is valid
// for the given identifier.
func validateNonce(identifier, nonce []byte) bool {
	h := sha256.New()
	h.Write(identifier)
	h.Write(nonce)
	ck := h.Sum(nil)
	return validateBitZeros(ck, global.NeedBits)
}
