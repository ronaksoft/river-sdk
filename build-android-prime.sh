#!/usr/bin/env bash

CGO_ENABLED=0 GO111MODULE=off gomobile bind -target=android -trimpath -ldflags="-s -w" -o="$RIVER_ANDROID_PATH"/river/sdk/riversdk.aar git.ronaksoft.com/river/sdk/sdk/prime
