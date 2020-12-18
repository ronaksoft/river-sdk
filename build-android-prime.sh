#!/usr/bin/env bash

CGO_ENABLED=0 GO111MODULE=on gomobile bind -target=android -trimpath -ldflags="-s -w" -o=$RIVER_ANDROID_PATH/river-module/sdk/riversdk.aar git.ronaksoft.com/river/sdk/sdk/prime
