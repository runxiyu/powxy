// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import "unsafe"

// Converts a string to a byte slice without copying the string.
// Memory is borrowed from the string.
// The resulting byte slice must not be modified in any form.
func stringToBytes(s string) (bytes []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
