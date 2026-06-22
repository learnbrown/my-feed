#!/bin/sh

go fmt ./...
go build -o build/main ./cmd/main.go
./build/main