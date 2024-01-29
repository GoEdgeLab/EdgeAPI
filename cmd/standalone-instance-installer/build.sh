#!/usr/bin/env bash

OS="${1}"
ARCH="${2}"
TAG="${3}"

if [ -z "$OS" ]; then
	echo "usage: build.sh OS ARCH"
	exit
fi

if [ -z "$ARCH" ]; then
	echo "usage: build.sh OS ARCH"
	exit
fi

env GOOS=linux GOARCH="${ARCH}" go build -tags="${TAG}" -trimpath -ldflags="-s -w" -o "standalone-instance-installer-${OS}-${ARCH}" main.go