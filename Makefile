SHELL := /bin/bash

test: lint vet build

lint:
	golangci-lint run

vet:
	go vet ./...

build: clean
	go build -v -o ./bin/traderui

clean:
	rm -rf ./bin
	rm -rf ./tmp
	rm -rf ./dist

# Commands for docker images.
# ----------------------------
build-linux:
	GOOS=linux GOARCH=amd64 go build -v -o ./bin/traderui .

build-docker: clean build-linux
	docker build -t quickfixgo/traderui:latest .