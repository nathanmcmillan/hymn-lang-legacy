#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh hymn_compiler -w out/hymn "$@"
