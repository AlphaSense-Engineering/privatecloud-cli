#!/usr/bin/env ash

set -euox pipefail

go mod tidy

go test -json ./...
