#!/bin/bash -e
cd "$(dirname "$0")"

if [ -f bin/hymn ]; then
  rm bin/hymn
fi
cd go
go build -o hymn
cd ..
mv go/hymn bin/hymn
