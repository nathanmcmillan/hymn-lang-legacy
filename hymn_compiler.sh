#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh hymn_compiler/hymn.hm -w out/hymn "$@"
