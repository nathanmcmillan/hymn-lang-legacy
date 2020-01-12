#!/bin/bash -e
cd "$(dirname "$0")"

cd go
go test -v -args $@
cd ..
