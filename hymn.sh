#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh

HYMN_PACKAGES=""
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/std"
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/books"
export HYMN_PACKAGES

if [ -f bin/hymn ]; then
  lib="$PWD/lib"
  path="$1"
  shift
  bin/hymn -d="$lib" -p="$path" "$@"
fi
