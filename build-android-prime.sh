#!/usr/bin/env bash

currentDir="$(pwd)"
dst="$GOPATH/src/git.ronaksoft.com/river/sdk"
echo "GOPATH = $GOPATH"
echo "PWD = $currentDir"
echo "DST = $dst"
echo "RIVER_ANDROID_PATH = $RIVER_ANDROID_PATH"

if [ "$currentDir" != "$dst" ]; then
  rm -r "$dst"
  mkdir -p "$dst"
  cp -r ./* "$dst"
fi
cd "$dst" || exit
echo "Switched current working directory to $(pwd)"
CGO_ENABLED=0 GO111MODULE=off gomobile bind -target=android -trimpath -ldflags="-s -w" -o="$RIVER_ANDROID_PATH"/river/sdk/riversdk.aar git.ronaksoft.com/river/sdk/sdk/prime
