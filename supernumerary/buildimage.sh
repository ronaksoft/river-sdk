#!/usr/bin/env bash
cd node/
go build -ldflags "-s -w" -o main
docker build -t supernumerary .
rm main

