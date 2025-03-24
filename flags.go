// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/sha256"
	"flag"
)

var (
	listenAddr        string
	destHost          string
	secondary         bool
	readTimeout       int
	writeTimeout      int
	idleTimeout       int
	readHeaderTimeout int
	ircAddr           string
	ircNet            string
	ircTLS            bool
	ircChannel        string
	ircNick           string
	ircUsername       string
	ircRealname       string
	ircBuf            uint
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
	flag.StringVar(&ircAddr, "irc-addr", "irc.runxiyu.org:6697", "irc server address")
	flag.StringVar(&ircNet, "irc-net", "tcp", "irc network transport")
	flag.BoolVar(&ircTLS, "irc-tls", true, "irc tls")
	flag.StringVar(&ircChannel, "irc-channel", "#logs", "irc channel")
	flag.StringVar(&ircNick, "irc-nick", "powxy", "irc nick")
	flag.StringVar(&ircUsername, "irc-username", "powxy", "irc username")
	flag.StringVar(&ircRealname, "irc-realname", "powxy", "irc realname")
	flag.UintVar(&ircBuf, "irc-buf", 3000, "irc buffer size")
	flag.Parse()
	global.NeedBitsReverse = sha256.Size - global.NeedBits
}
