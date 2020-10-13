#!/usr/bin/env bash

function build() {
	ROOT=$(dirname $0)
	NAME="edge-api"
	DIST="../dist/${NAME}"
	OS=${1}
	ARCH=${2}

	if [ -z $OS ]; then
		echo "usage: build.sh OS ARCH"
		exit
	fi
	if [ -z $ARCH ]; then
		echo "usage: build.sh OS ARCH"
		exit
	fi

	VERSION_DATA=$(cat ../internal/const/const.go)
	re="Version[ ]+=[ ]+\"([0-9.]+)\""
	if [[ $VERSION_DATA =~ $re ]]; then
		VERSION=${BASH_REMATCH[1]}
	else
		echo "could not match version"
		exit
	fi

	ZIP="${NAME}-${OS}-${ARCH}-v${VERSION}.zip"

	# copy files
	echo "copying ..."
	if [ ! -d $DIST ]; then
		mkdir $DIST
		mkdir $DIST/bin
		mkdir $DIST/configs
		mkdir $DIST/logs
	fi
	cp configs/api.template.yaml $DIST/configs/
	cp configs/db.template.yaml $DIST/configs/
	cp -R deploy $DIST/
	cp -R installers $DIST/

	# building installer
	echo "building installer ..."
	architects=("amd64" "386")
	for arch in "${architects[@]}"; do
		# TODO support arm, mips ...
		env GOOS=linux GOARCH=${arch} go build --ldflags="-s -w" -o $ROOT/installers/installer-helper-linux-${arch} $ROOT/../cmd/installer-helper/main.go
	done

	# building api node
	env GOOS=$OS GOARCH=$ARCH go build --ldflags="-s -w" -o $DIST/bin/edge-api $ROOT/../cmd/edge-api/main.go

	echo "zip files"
	cd "${DIST}/../" || exit
	if [ -f "${ZIP}" ]; then
		rm -f "${ZIP}"
	fi
	zip -r -X -q "${ZIP}" ${NAME}/
	rm -rf ${NAME}
	cd - || exit

	echo "[done]"
}

build $1 $2
