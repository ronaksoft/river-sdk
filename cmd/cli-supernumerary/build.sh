#!/usr/bin/env bash
cd node/
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o ./_build/supernumerary
docker build -t registry.ronaksoftware.com/ronak/riversdk/supernumerary:0.1 .

