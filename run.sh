#!/bin/sh

set -e

go fmt ./...

go build -o build/main ./cmd/main.go

# ./run.sh -mysql
if [ "$1" = "-mysql" ]; then
    DB_DRIVER=mysql MYSQL_USER=dev_user MYSQL_PASSWORD=qwerdf MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_DATABASE=myfeed ./build/main
else
    ./build/main
fi