#!/usr/bin/env bash

function build() {
	ROOT=$(dirname "$0")

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

	VERSION=$(lookup_version "${ROOT}/../../internal/const/const.go")
	TARGET_NAME="edge-instance-installer-${OS}-${ARCH}-v${VERSION}"

	env GOOS=linux GOARCH="${ARCH}" go build -tags="${TAG}" -trimpath -ldflags="-s -w" -o "${TARGET_NAME}" main.go

	if [ -f "${TARGET_NAME}" ]; then
		cp "${TARGET_NAME}" "${ROOT}/../../../EdgeAdmin/docker/instance/edge-instance/assets"
	fi

	echo "[done]"
}

function lookup_version() {
	FILE=$1
	VERSION_DATA=$(cat "$FILE")
	re="Version[ ]+=[ ]+\"([0-9.]+)\""
	if [[ $VERSION_DATA =~ $re ]]; then
		VERSION=${BASH_REMATCH[1]}
		echo "$VERSION"
	else
		echo "could not match version"
		exit
	fi
}

build "$1" "$2" "$3"