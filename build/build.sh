#!/usr/bin/env bash

ROOT=`dirname $0`

# building installer
architects=( "amd64" "386" )
for arch in "${architects[@]}"
do
	# TODO support arm, mips ...
	env GOOS=linux GOARCH=${arch} go build --ldflags="-s -w" -o $ROOT/installers/installer-helper-linux-${arch} $ROOT/../cmd/installer-helper/main.go
done
