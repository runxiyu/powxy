// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var reverseProxy *httputil.ReverseProxy

// This init sets up the reverse proxy. Go's NewSingleHostReverseProxy is
// sufficient for our use case.
func init() {
	parsedURL, err := url.Parse(destHost)
	if err != nil {
		log.Fatal(err)
	}
	reverseProxy = httputil.NewSingleHostReverseProxy(parsedURL)
}

// proxyRequest proxies the incoming request to the destination host.
func proxyRequest(writer http.ResponseWriter, request *http.Request) {
	reverseProxy.ServeHTTP(writer, request)
}
