#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh test_hymn_std/test_path.hm -t -w out/test_hymn_std "$@"
