#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh boot/main.hm -w out/hymn -v "std=$(pwd)/std"
