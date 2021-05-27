##!/usr/bin/env bash

currentDir="$(pwd)"
dst="$GOPATH/src/git.ronaksoft.com/river/sdk"
echo "GOPATH = $GOPATH"
echo "PWD = $currentDir"
echo "DST = $dst"
echo "RIVER_IOS_PATH = $RIVER_IOS_PATH"

if [ "$currentDir" != "$dst" ]; then
  mkdir -p "$dst"
  cp -r ./* "$dst"
fi
cd "$dst" || exit
echo "Switched current working directory to $(pwd)"
CGO_ENABLED=0 GO111MODULE=off gomobile bind -target=ios -trimpath -o="$RIVER_IOS_PATH"/riversdk.framework git.ronaksoft.com/river/sdk/sdk/prime