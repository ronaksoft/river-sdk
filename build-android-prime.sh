#!/usr/bin/env bash

cp -r ./* "$GOPATH"/src/git.ronaksoft.com/river/sdk/
CGO_ENABLED=0 GO111MODULE=on gomobile bind -target=android -trimpath -ldflags="-s -w" -o="$RIVER_ANDROID_PATH"/river/sdk/riversdk.aar git.ronaksoft.com/river/sdk/sdk/prime
