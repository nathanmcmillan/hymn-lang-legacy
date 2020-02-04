#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh

export HYMN_MODULES="root=$(pwd)"

if [ -f bin/hymn ]; then
  lib="$PWD/lib"
  path="$1"
  shift
  bin/hymn -d="$lib" -p="$path" "$@"
fi
