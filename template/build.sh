#!/bin/sh

mkdir -p build
cp -R ./migration ./build
cp -R ./script ./build
cp -R ./config ./build
cp cmd/{{app_name2}}_config.ini ./build

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./build/{{app_name2}} ./cmd
