#!/usr/bin/env bash

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
go get -u git.ronaksoftware.com/river/msg
go mod vendor