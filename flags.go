package main

import "flag"

var (
	difficulty uint
	listenAddr string
	destHost   string
	secondary  bool
)

func init() {
	flag.UintVar(&difficulty, "difficulty", 17, "leading zero bits required for the challenge")
	flag.StringVar(&listenAddr, "listen", ":8081", "address to listen on")
	flag.StringVar(&destHost, "upstream", "http://127.0.0.1:8080", "destination url base to proxy to")
	flag.BoolVar(&secondary, "secondary", false, "trust X-Forwarded-For headers")
	flag.Parse()
}
