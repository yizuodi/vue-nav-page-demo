#!/usr/bin/env bash
set -euo pipefail
mkdir -p dist
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o dist/nav-server-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o dist/nav-server-windows-amd64.exe main.go
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o dist/nav-server-darwin-arm64 main.go
ls -lh dist
