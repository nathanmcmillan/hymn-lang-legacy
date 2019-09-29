#!/bin/bash -e
cd "$(dirname "$0")"

path="$HOME/hymn"

if [ -e "$path" ]; then
    echo "hymn is already installed at $path"
    exit
fi

mkdir -p "$path"
cd "$path"
git clone -b stable https://github.com/gameinbucket/hymn-lang.git .

./make.sh

echo "todo export PATH"
export PATH="$PATH:$HOME/hymn/bin"

echo "hymn installed successfully"
