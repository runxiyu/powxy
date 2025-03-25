// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Vicky Williams <https://que.trolling.win>
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

#include "sha256.h"

unsigned char challenge[32];

char validate_hash(unsigned char *hash, unsigned char zero_bit_count)
{
	unsigned char q = zero_bit_count / 8;
	unsigned char r = zero_bit_count % 8;

	for (unsigned char i = 0; i < q; i++)
		if (hash[i] != 0)
			return 0;
	if (r > 0) {
		unsigned char mask = (unsigned char)(0xFF << (8 - r));
		if (hash[q] & mask)
			return 0;
	}

	return 1;
}

unsigned char *get_challenge_ptr()
{
	return challenge;
}

unsigned long long solve(unsigned char difficulty)
{
	unsigned long long nonce;
	SHA256_CTX context;

	unsigned char hash[32];

	nonce = 0;

	for (;;) {
		sha256_init(&context);
		sha256_update(&context, challenge, sizeof(challenge));
		sha256_update(&context, (unsigned char *)(&nonce),
			      sizeof(nonce));
		sha256_final(&context, hash);

		if (validate_hash(hash, difficulty))
			break;

		nonce++;
	}

	return nonce;
}
