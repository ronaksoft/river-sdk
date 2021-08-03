##!/usr/bin/env bash

currentDir="$(pwd)"
dst="$GOPATH/src/git.ronaksoft.com/river/sdk"
echo "GOPATH = $GOPATH"
echo "PWD = $currentDir"
echo "DST = $dst"
echo "RIVER_IOS_PATH = $RIVER_IOS_PATH"

if [ "$currentDir" != "$dst" ]; then
  rm -r "$dst"
  mkdir -p "$dst"
  cp -r ./* "$dst"
fi
cd "$dst" || exit
echo "Switched current working directory to $(pwd)"
CGO_ENABLED=0 GO111MODULE=off gomobile bind -target=ios -o="$RIVER_IOS_PATH"/minisdk.framework git.ronaksoft.com/river/sdk/sdk/mini