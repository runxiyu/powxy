package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var reverseProxy *httputil.ReverseProxy

func init() {
	parsedURL, err := url.Parse(destHost)
	if err != nil {
		log.Fatal(err)
	}
	reverseProxy = httputil.NewSingleHostReverseProxy(parsedURL)
}

func proxyRequest(writer http.ResponseWriter, request *http.Request) {
	reverseProxy.ServeHTTP(writer, request)
}
