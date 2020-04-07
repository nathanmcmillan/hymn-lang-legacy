#!/bin/bash -e
cd "$(dirname "$0")"

# read -p "Where do you want to install Hymn? " answer

base="$HOME/.hymn"

if [ -e "$base" ]; then
    echo "Hymn is already installed at $base"
    # read -p "Hymn is already installed here. Do you want to overwrite it? " answer
    exit 1
fi

./make.sh

mkdir -p "$base"
cp -r bin "$base"
cp -r hymn_std "$base"
cp -r libc "$base"

echo 'Success!'
echo 'Please add '$base'/bin to your $PATH variable'

# echo 'Please install a C compiler such as GCC or Clang'
