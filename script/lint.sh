#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "Running golangci-lint with verbose output (timeout: 3m)..."

golangci-lint run --verbose --timeout 3m0s
