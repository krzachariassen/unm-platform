.PHONY: test build lint tidy

test:
	go test ./...

test-v:
	go test ./... -v

test-cover:
	go test ./... -cover

build-server:
	go build ./cmd/server/

build-cli:
	go build ./cmd/cli/

build: build-server build-cli

lint:
	go vet ./...

tidy:
	go mod tidy

run-server:
	go run ./cmd/server/

run-cli:
	go run ./cmd/cli/ $(ARGS)
