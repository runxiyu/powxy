// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

// validateBitZeros checks if the first n bits of a byte slice are all zeros.
func validateBitZeros(bs []byte, n uint) bool {
	q := n / 8
	r := n % 8

	for i := range q {
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
