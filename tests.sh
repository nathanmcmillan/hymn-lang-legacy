#!/bin/bash -e
cd "$(dirname "$0")"

. ./hymn_packages.sh
. ./hymn_libc.sh

cd src
go test -v -args $@
cd ..
