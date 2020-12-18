#!/bin/bash

docker build --pull -t riversdk/android-builder .
docker run riversdk/android-builder