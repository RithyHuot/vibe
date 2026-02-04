#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "Scanning for known vulnerabilities in dependencies..."

# Run govulncheck on source code
# Display options can be controlled via VULNCHECK_SHOW env var
# Valid values: traces, color, version, verbose (comma-separated)
# Default: verbose (shows detailed output)
SHOW_FLAGS=${VULNCHECK_SHOW:-"verbose"}
go run golang.org/x/vuln/cmd/govulncheck -show "${SHOW_FLAGS}" ./...
