.DEFAULT_GOAL := test

.PHONY := build
build:
	go build ./...

.PHONY := test
test:
	go test ./...
