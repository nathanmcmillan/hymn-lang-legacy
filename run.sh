#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh

if [ -f bin/hymn ]; then
  lib="$PWD/lib"
  bin/hymn "build" $lib $@
fi
