#!/usr/bin/env bash
cd node/
env CGO_ENABLED=0 go build -ldflags "-s -w" -o main
docker build -t supernumerary .
rm main

