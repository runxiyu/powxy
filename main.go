package main

import (
	"io"
	"log"
	"maps"
	"net/http"
)

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
}

func main() {
	log.Fatal(http.ListenAndServe(":8081", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.RemoteAddr, request.RequestURI)

		request.Host = "127.0.0.1:8080"
		request.URL.Host = "127.0.0.1:8080"
		request.URL.Scheme = "http"
		request.RequestURI = ""

		response, err := client.Do(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadGateway)
			return
		}

		maps.Copy(writer.Header(), response.Header)
		writer.WriteHeader(response.StatusCode)
		io.Copy(writer, response.Body)
	})))
}
