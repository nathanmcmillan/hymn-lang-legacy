#!/bin/bash -e
cd "$(dirname "$0")"

HYMN_PACKAGES="$HYMN_PACKAGES:$(pwd)/boot"
export HYMN_PACKAGES

./hymn.sh test_boot/test_tokenizer.hm -t -w out/test_hymn "$@"
