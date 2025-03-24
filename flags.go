// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import "flag"

var (
	listenAddr        string
	destHost          string
	secondary         bool
	readTimeout       int
	writeTimeout      int
	idleTimeout       int
	readHeaderTimeout int
)

// This init parses command line flags.
func init() {
	flag.UintVar(&global.NeedBits, "difficulty", 20, "leading zero bits required for the challenge")
	flag.StringVar(&global.SourceURL, "source", "https://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/", "url to the source code")
	flag.StringVar(&listenAddr, "listen", ":8081", "address to listen on")
	flag.StringVar(&destHost, "upstream", "http://127.0.0.1:8080", "destination url base to proxy to")
	flag.BoolVar(&secondary, "secondary", false, "trust X-Forwarded-For headers")
	flag.IntVar(&readTimeout, "read-timeout", 0, "read timeout in seconds, 0 for no timeout")
	flag.IntVar(&writeTimeout, "write-timeout", 0, "write timeout in seconds, 0 for no timeout")
	flag.IntVar(&idleTimeout, "idle-timeout", 0, "idle timeout in seconds, 0 for no timeout")
	flag.IntVar(&readHeaderTimeout, "read-header-timeout", 30, "read header timeout in seconds, 0 for no timeout")
	flag.Parse()
	global.NeedBitsReverse = 256 - global.NeedBits
}
