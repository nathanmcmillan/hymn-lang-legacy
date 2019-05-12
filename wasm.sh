cd c-code
source $HOME/Documents/programming/online/emsdk/emsdk_env.sh --build=Release
emcc ss.c -s WASM=1 -o ss.html
emrun --no_browser --port 3000 .
