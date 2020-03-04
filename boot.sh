#!/bin/bash -e
cd "$(dirname "$0")"

./hymn.sh boot/main.hm -w out -v "std=$(pwd)/std"
