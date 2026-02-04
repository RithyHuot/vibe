#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "Formatting Go imports using goimports-reviser..."
echo "  - Removing unused imports"
echo "  - Setting import aliases"
echo "  - Organizing imports (stdlib → third-party → local)"

# Get the module name from go.mod
MODULE_NAME=$(go list -m)

# Run goimports-reviser with recommended flags
go run github.com/incu6us/goimports-reviser/v3 \
  -rm-unused \
  -set-alias \
  -format \
  -recursive \
  -project-name="${MODULE_NAME}" \
  ./...
