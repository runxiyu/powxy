// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"testing"
)

func TestValidateBitZeros(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		n        uint
		expected bool
	}{
		{
			name:     "First 8 bits are zeros",
			bs:       []byte{0x00, 0x01},
			n:        8,
			expected: true,
		},
		{
			name:     "First 8 bits are not all zeros",
			bs:       []byte{0x01, 0x00},
			n:        8,
			expected: false,
		},
		{
			name:     "First 16 bits are zeros",
			bs:       []byte{0x00, 0x00, 0x01},
			n:        16,
			expected: true,
		},
		{
			name:     "First 16 bits are not all zeros",
			bs:       []byte{0x01, 0x00, 0x00},
			n:        16,
			expected: false,
		},
		{
			name:     "First 9 bits are zeros",
			bs:       []byte{0x00, 0x01},
			n:        9,
			expected: true,
		},
		{
			name:     "First 9 bits are not all zeros",
			bs:       []byte{0x01, 0x01},
			n:        9,
			expected: false,
		},
		{
			name:     "First 10 bits are zeros",
			bs:       []byte{0x00, 0x20},
			n:        10,
			expected: true,
		},
		{
			name:     "First 10 bits are not all zeros",
			bs:       []byte{0x00, 0x40},
			n:        10,
			expected: false,
		},
		{
			name:     "First 24 bits are zeros",
			bs:       []byte{0x00, 0x00, 0x00, 0x01},
			n:        24,
			expected: true,
		},
		{
			name:     "First 24 bits are not all zeros",
			bs:       []byte{0x00, 0x01, 0x00, 0x00},
			n:        24,
			expected: false,
		},
		{
			name:     "Checking zero bits",
			bs:       []byte{0xFF, 0xFF},
			n:        0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateBitZeros(tt.bs, tt.n)
			if got != tt.expected {
				t.Errorf("validateBitZeros(%v, %v) = %v; want %v", tt.bs, tt.n, got, tt.expected)
			}
		})
	}
}
