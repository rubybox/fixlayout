#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
command -v go >/dev/null || { echo 'Please install Go (sudo apt install golang-go)' >&2; exit 1; }
command -v gcc >/dev/null || { echo 'Please install GCC (sudo apt install build-essential)' >&2; exit 1; }
mkdir -p bin
CGO_ENABLED=1 go build -trimpath -ldflags='-s -w' -o bin/fixlayout-go .
echo "Built: $(pwd)/bin/fixlayout-go"
