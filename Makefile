powxy: *.go version.go wasm/solver.wasm
	go build -o powxy

version.go:
	printf 'package main\n\nfunc init() {\n\tglobal.Version = "%s"\n}\n' $(shell git describe --tags --always --dirty) > $@

wasm/solver.wasm: wasm/solver.c wasm/sha256.c wasm/sha256.h
	clang --target=wasm32 -nostdlib -Wl,--no-entry -Wl,--export-all -o wasm/solver.wasm wasm/solver.c wasm/sha256.c

.PHONY: version.go
