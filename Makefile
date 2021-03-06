#!/usr/bin/make -f

# make is funny

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
ldflags := -ldflags="-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

export GO111MODULE=on
export GOPROXY=https://goproxy.io,direct

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -v $(ldflags) -o build/github-console-linux .

.PHONY: go.mod
go.mod:
	@$(PROXY) go mod tidy
	@$(PROXY) go mod download
	@$(PROXY) go mod verify

.PHONY: install
install:
	@go install -v $(ldflags) .
