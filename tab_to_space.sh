#!/bin/bash
set -e
set -u

exception() {
    echo "file?"
    exit 1
}

if [ $# -lt 1 ]; then
    exception
fi

if [ ! -f "$1" ]; then
    exception
fi

temp=$(mktemp)
expand --tabs=4 "$1" > "$temp"
mv "$temp" "$1"
