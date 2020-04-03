#!/bin/bash -e
cd "$(dirname "$0")"

if [ -f bin/hymn ]; then
  rm bin/hymn
fi
cd src
go build -o hymn
cd ..
if [ -f src/hymn ]; then
  if [ ! -d bin ]; then
    mkdir -p bin
  fi
  mv src/hymn bin/hymn
fi
