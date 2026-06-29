#!/bin/sh

set -e

go fmt ./...

go run ./cmd/main.go