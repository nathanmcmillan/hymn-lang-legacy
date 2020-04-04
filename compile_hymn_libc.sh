#!/bin/sh -e
cd "$(dirname "$0")"
cd libc
gcc \
-Wall \
-Wextra \
-Werror \
-pedantic \
-std=c11 \
hmlib_string.c \
hmlib_files.c \
hmlib_slice.c \
hmlib_system.c \
-c

mv *.o link/.
