#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh hymn_core/file.hm -x -w core_c "$@"
