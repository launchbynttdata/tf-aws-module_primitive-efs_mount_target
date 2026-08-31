#!/usr/bin/env bash
# Run golangci-lint via go.mod's tool directive (same path as Makefile lint).
set -euo pipefail

go_test_directories="${GO_TEST_DIRECTORIES:-tests}"

if ! find "$go_test_directories" -name '*.go' -type f 2>/dev/null | grep -q '\.go'; then
  exit 0
fi

exec go tool golangci-lint run -c .golangci.yaml --fix "./${go_test_directories}/..."
