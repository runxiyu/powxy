// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

const solverProgram = `// You need to have OpenSSL, and link with -lcrypto

#include <openssl/evp.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <stdlib.h>
#include <errno.h>

bool validate_bit_zeros(const unsigned char *bs, uint8_t n)
{
	uint8_t q = n / 8;
	uint8_t r = n % 8;

	for (uint8_t i = 0; i < q; i++) {
		if (bs[i] != 0)
			return false;
	}

	if (r > 0) {
		unsigned char mask = (unsigned char)(0xFF << (8 - r));
		if (bs[q] & mask)
			return false;
	}

	return true;
}

int main(int argc, char **argv)
{
	if (argc < 3) {
		fprintf(stderr, "usage: %s <base64_data> <difficulty>\n",
			argv[0]);
		return 1;
	}

	size_t base64_data_len = strlen(argv[1]);
	unsigned char *base64_data = malloc(base64_data_len);
	if (!base64_data) {
		perror("malloc");
		return 1;
	}
	memcpy(base64_data, argv[1], base64_data_len);

	char *endptr = NULL;
	errno = 0;
	unsigned long tmp_val = strtoul(argv[2], &endptr, 10);
	if ((errno == ERANGE && tmp_val == ULONG_MAX) || *endptr != '\0'
	    || tmp_val > UINT8_MAX) {
		fprintf(stderr, "invalid difficulty value\n");
		free(base64_data);
		return 1;
	}
	uint8_t difficulty = (uint8_t) tmp_val;

	BIO *b64 = BIO_new(BIO_f_base64());
	BIO *bmem = BIO_new_mem_buf(base64_data, (int)base64_data_len);
	if (!b64 || !bmem) {
		fprintf(stderr, "BIO_new/BIO_new_mem_buf\n");
		free(base64_data);
		return 1;
	}

	BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
	b64 = BIO_push(b64, bmem);

	size_t decoded_cap = base64_data_len;
	unsigned char *decoded = malloc(decoded_cap);
	if (!decoded) {
		perror("malloc");
		BIO_free_all(b64);
		free(base64_data);
		return 1;
	}

	int decoded_len = BIO_read(b64, decoded, (int)decoded_cap);
	if (decoded_len < 0) {
		fprintf(stderr, "BIO_read\n");
		BIO_free_all(b64);
		free(base64_data);
		free(decoded);
		return 1;
	}
	BIO_free_all(b64);
	free(base64_data);

	EVP_MD_CTX *mdctx = EVP_MD_CTX_new();
	if (!mdctx) {
		fprintf(stderr, "EVP_MD_CTX_new\n");
		free(decoded);
		return 1;
	}

	size_t len = EVP_MD_size(EVP_sha256());
	unsigned char digest[EVP_MAX_MD_SIZE];
	size_t next = 0;

	while (1) {
		if (EVP_DigestInit_ex(mdctx, EVP_sha256(), NULL) != 1) {
			fprintf(stderr, "EVP_DigestInit_ex\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestUpdate(mdctx, decoded, decoded_len) != 1) {
			fprintf(stderr, "EVP_DigestUpdate(data)\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestUpdate(mdctx, &next, sizeof(next)) != 1) {
			fprintf(stderr, "EVP_DigestUpdate(next)\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (EVP_DigestFinal_ex(mdctx, digest, NULL) != 1) {
			fprintf(stderr, "EVP_DigestFinal_ex\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
		if (validate_bit_zeros(digest, difficulty)) {
			break;
		}
		next++;
		if (!next) {
			fprintf(stderr, "unsigned integer overflow\n");
			EVP_MD_CTX_free(mdctx);
			free(decoded);
			return 1;
		}
	}
	EVP_MD_CTX_free(mdctx);
	free(decoded);

	BIO *b64_out = BIO_new(BIO_f_base64());
	BIO *bmem_out = BIO_new(BIO_s_mem());
	if (!b64_out || !bmem_out) {
		fprintf(stderr, "BIO_new\n");
		if (b64_out)
			BIO_free_all(b64_out);
		if (bmem_out)
			BIO_free(bmem_out);
		return 1;
	}
	BIO_set_flags(b64_out, BIO_FLAGS_BASE64_NO_NL);
	b64_out = BIO_push(b64_out, bmem_out);

	if (BIO_write(b64_out, &next, sizeof(next)) < 0) {
		fprintf(stderr, "BIO_write\n");
		BIO_free_all(b64_out);
		return 1;
	}
	if (BIO_flush(b64_out) < 1) {
		fprintf(stderr, "BIO_flush\n");
		BIO_free_all(b64_out);
		return 1;
	}

	BUF_MEM *bptr = NULL;
	BIO_get_mem_ptr(b64_out, &bptr);
	if (!bptr || !bptr->data) {
		fprintf(stderr, "BIO_get_mem_ptr\n");
		BIO_free_all(b64_out);
		return 1;
	}

	write(STDOUT_FILENO, bptr->data, bptr->length);
	write(STDOUT_FILENO, "\n", 1);

	BIO_free_all(b64_out);
	return 0;
}`
