.PHONY: build test lint install clean run

# Go toolchain
export GOTOOLCHAIN=auto

BINARY_NAME=vibe
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/vibe/main.go

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./... || true
	@if [ -f coverage.out ]; then go tool cover -html=coverage.out; fi

coverage: test-coverage

lint:
	golangci-lint run

install:
	go install ${LDFLAGS} ./cmd/vibe

clean:
	rm -rf bin/ coverage.out

run:
	go run cmd/vibe/main.go

deps:
	go mod download
	go mod tidy
