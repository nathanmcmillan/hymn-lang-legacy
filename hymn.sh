#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh

if [ ! -f bin/hymn ]; then
    exit 1
fi

HYMN_PACKAGES="${HYMN_PACKAGES:-''}"
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/hymn"
HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/books"
export HYMN_PACKAGES

HYMN_LIBC="$(pwd)/libc"
export HYMN_LIBC

path="$1"
shift
bin/hymn -p="$path" "$@"
