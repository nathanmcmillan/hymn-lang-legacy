#!/bin/bash -e
cd "$(dirname "$0")"

apt-get update
apt-get install gcc
apt-get install clang-tools
apt-get install golang
