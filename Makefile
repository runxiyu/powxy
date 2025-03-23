powxy: *.go version.go
	go build -o powxy

version.go:
	printf 'package main\n\nfunc init() {\n\tglobal.Version = "%s"\n}\n' $(shell git describe --tags --always --dirty) > $@

.PHONY: version.go
