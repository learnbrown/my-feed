#!/bin/sh

go fmt ./...
go build ./cmd/main.go
./main