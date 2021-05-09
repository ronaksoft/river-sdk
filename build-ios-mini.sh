##!/usr/bin/env bash

CGO_ENABLED=0 gomobile bind -target=ios -o=$RIVER_IOS_PATH/minisdk.framework git.ronaksoft.com/river/sdk/sdk/mini