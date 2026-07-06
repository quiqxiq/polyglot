.PHONY: build test lint fmt run

build:
	go build ./...

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

fmt:
	gofmt -l .
	gofumpt -l . || true

run:
	go run ./cmd/server
