#!/usr/bin/env bash

#go get -u git.ronaksoft.com/river/msg
go mod vendor
go generate ./...
go vet ./...
go fmt ./...