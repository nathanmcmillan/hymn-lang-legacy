#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh boot_tests -w out/hymn_test "$@"
