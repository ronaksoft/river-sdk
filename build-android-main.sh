#!/usr/bin/env bash

gomobile bind -target=android -ldflags="-s -w" -o=$RIVER_ANDROID_PATH/mainmodule/sdk/riversdk.aar git.ronaksoft.com/river/sdk/sdk/main
