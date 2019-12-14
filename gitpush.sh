#!/usr/bin/env bash

go generate ./...
go vet ./...
git add .
git commit -m "$1"
git push --tags