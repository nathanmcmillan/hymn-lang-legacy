#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh hymn_core/hymn.hm -w hymn_core/c "$@"
