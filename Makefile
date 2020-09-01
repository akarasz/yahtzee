.DEFAULT_GOAL := run

.PHONY := build
build:
	go build ./...

.PHONY := test
test:
	go test --count=1 ./...

.PHONY := run
run:
	go run cmd/server/main.go
