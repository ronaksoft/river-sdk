##!/usr/bin/env bash

CGO_ENABLED=0 gomobile bind -target=ios -trimpath -o=$RIVER_IOS_PATH/riversdk.framework git.ronaksoft.com/river/sdk/sdk/prime