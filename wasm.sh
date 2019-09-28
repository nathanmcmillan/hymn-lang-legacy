#!/bin/bash -e
cd "$(dirname "$0")"

source $HOME/Documents/programming/online/emsdk/emsdk_env.sh --build=Release
emcc "$@" -s WASM=1 -o wasm-test.html
emrun --no_browser --port 3000 .
