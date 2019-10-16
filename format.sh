#!/bin/bash -e
cd "$(dirname "$0")"

./make.sh && bin/hymn fmt "$1"
