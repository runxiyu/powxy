// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

func validateBitZeros(bs []byte, n int) bool {
	q := n / 8
	r := n % 8

	if !bytes.Equal(bs[:q], make([]byte, q)) {
		return false
	}
	if r != 0 && (bs[q]&(0xFF<<(8-r)) != 0) {
		return false
	}
	return true
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: program <base64 input> <difficulty>")
		os.Exit(1)
	}

	decoded, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid base64:", err)
		os.Exit(1)
	}

	difficulty, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid difficulty:", err)
		os.Exit(1)
	}

	numWorkers := runtime.NumCPU()
	result := make(chan uint64)

	for i := 0; i < numWorkers; i++ {
		go func(start, step int) {
			var nextVal uint64 = uint64(start)

			buf := append(decoded, make([]byte, 8)...)

			for {
				binary.LittleEndian.PutUint64(buf[len(buf)-8:], nextVal)
				h := sha256.Sum256(buf)

				if validateBitZeros(h[:], difficulty) {
					result <- nextVal
					return
				}

				nextVal = (nextVal + uint64(step)) & 0xFFFFFFFFFFFFFFFF
				if nextVal == uint64(start) {
					fmt.Fprintln(os.Stderr, "overflow")
					os.Exit(1)
				}
			}
		}(i, numWorkers)
	}

	found := <-result
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, found)
	fmt.Println(base64.StdEncoding.EncodeToString(out))
}
