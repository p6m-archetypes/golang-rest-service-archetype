.PHONY: setup build test lint run

setup:
	go mod tidy

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run

run:
	go run ./cmd/server
