#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh

if [ ! -f bin/hymn ]; then
    exit 1
fi

. ./hymn_packages.sh
. ./hymn_libc.sh

path="$1"
shift
bin/hymn -p="$path" "$@"
