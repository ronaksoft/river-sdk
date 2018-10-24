#!/usr/bin/env bash

go build -ldflags "-s -w" -o main *.go
#go build -o $GOPATH/bin/river-cli
