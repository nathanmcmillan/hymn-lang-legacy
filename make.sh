#!/bin/bash -e
cd "$(dirname "$0")"

if [ -f bin/hymn ]; then
  rm bin/hymn
fi
cd go
go build -o hymn
cd ..
if [ -f go/hymn ]; then
  if [ ! -d bin ]; then
    mkdir -p bin
  fi
  mv go/hymn bin/hymn
fi
