#!/bin/bash

HYMN_STD="$(pwd)/hymn_std"
HYMN_BOOKS="$(pwd)/books"
HYMN_PACKAGES="{\"hymn\":\"${HYMN_STD}\",\"books\":\"${HYMN_BOOKS}\"}"
export HYMN_PACKAGES
