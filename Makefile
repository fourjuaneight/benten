VERSION := 1.0.0

.PHONY: run build install

run: benten.go go.mod
	chmod +x benten.go
	cd "${GOPATH}" && go install github.com/erning/gorun@latest

# https://golang.org/cmd/link/
build: benten.go go.mod
	go build -o benten benten.go

install: benten.go go.mod
	go install ./benten.go
	cp .env ${GOPATH}/.env
