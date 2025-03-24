# SPDX-License-Identifier: BSD-2-Clause
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

import base64
import hashlib
import sys
import struct

def validate_bit_zeros(bs: bytes, n: int) -> bool:
	q, r = divmod(n, 8)
	if any(b != 0 for b in bs[:q]):
		return False
	if r and (bs[q] & (0xFF << (8 - r))):
		return False
	return True

decoded = base64.b64decode(sys.argv[1])
difficulty = int(sys.argv[2])
next_val = 0

while True:
	h = hashlib.sha256(decoded + struct.pack("Q", next_val)).digest()
	if validate_bit_zeros(h, difficulty):
		break
	next_val = (next_val + 1) & 0xFFFFFFFFFFFFFFFF
	if next_val == 0:
        raise ValueError("overflow")

print(base64.b64encode(struct.pack("Q", next_val)).decode())
