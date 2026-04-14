#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${GOROOT:-}" ]]; then
  export PATH="$GOROOT/bin:$PATH"
fi

export GOCACHE="${GOCACHE:-/tmp/go-build-worthly-tracker}"
export GOMODCACHE="${GOMODCACHE:-/tmp/go-mod-worthly-tracker}"

go get -u ./...
go mod tidy
