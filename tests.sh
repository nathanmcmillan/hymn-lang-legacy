#!/bin/bash -e
cd "$(dirname "$0")"

HYMN_PACKAGES=""
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/hymn"
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/books"
export HYMN_PACKAGES

HYMN_LIBC="$(pwd)/libc"
export HYMN_LIBC

cd go
go test -v -args $@
cd ..
