.PHONY: help build test lint install clean run fmt fmt-check vulncheck check pre-pr

# Go toolchain
export GOTOOLCHAIN=auto

BINARY_NAME=vibe
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the vibe binary
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/vibe/main.go

test: ## Run all tests with race detection
	go test -v -race ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./... || true
	@if [ -f coverage.out ]; then go tool cover -html=coverage.out; fi

coverage: test-coverage ## Alias for test-coverage

lint: ## Run golangci-lint
	@script/lint.sh

fmt: ## Format Go imports
	@script/fmt.sh

fmt-check: ## Check if imports are properly formatted
	@echo "Checking import formatting..."
	@git diff --exit-code -- '*.go' '**/*.go' || (echo "Uncommitted Go file changes detected. Commit or stash before running fmt-check." && exit 1)
	@script/fmt.sh
	@git diff --exit-code -- '*.go' '**/*.go' || (echo "Formatting changes detected. Run 'make fmt' and commit changes." && exit 1)

vulncheck: ## Scan for security vulnerabilities
	@script/vulncheck.sh

check: lint fmt-check test vulncheck ## Run all checks (lint, format, test, security)
	@echo "All checks passed!"

pre-pr: check ## Run all pre-PR checks (recommended before creating PR)
	@echo "✓ Pre-PR checks completed successfully!"
	@echo "Ready to create a pull request."

install: ## Install vibe to $GOPATH/bin
	go install ${LDFLAGS} ./cmd/vibe

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out

run: ## Run vibe locally
	go run cmd/vibe/main.go

deps: ## Download and tidy dependencies
	go mod download
	go mod tidy
