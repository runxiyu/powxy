// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import "unsafe"

func stringToBytes(s string) (bytes []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
