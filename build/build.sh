#!/usr/bin/env bash

function build() {
	ROOT=$(dirname "$0")
	NAME="edge-api"
	DIST=$ROOT/"../dist/${NAME}"
	OS=${1}
	ARCH=${2}
	TAG=${3}
	NODE_ARCHITECTS=("amd64" "arm64")

	if [ -z "$OS" ]; then
		echo "usage: build.sh OS ARCH"
		exit
	fi
	if [ -z "$ARCH" ]; then
		echo "usage: build.sh OS ARCH"
		exit
	fi
	if [ -z "$TAG" ]; then
		TAG="community"
	fi

	VERSION=$(lookup-version "$ROOT"/../internal/const/const.go)
	ZIP="${NAME}-${OS}-${ARCH}-${TAG}-v${VERSION}.zip"

	# build edge-node
	NodeVersion=$(lookup-version "$ROOT""/../../EdgeNode/internal/const/const.go")
	echo "building edge-node v${NodeVersion} ..."
	EDGE_NODE_BUILD_SCRIPT=$ROOT"/../../EdgeNode/build/build.sh"
	if [ ! -f "$EDGE_NODE_BUILD_SCRIPT" ]; then
		echo "unable to find edge-node build script 'EdgeNode/build/build.sh'"
		exit
	fi
	cd "$ROOT""/../../EdgeNode/build" || exit
	echo "=============================="
	for arch in "${NODE_ARCHITECTS[@]}"; do
		if [ ! -f "$ROOT""/../../EdgeNode/dist/edge-node-linux-${arch}-${TAG}-v${NodeVersion}.zip" ]; then
			./build.sh linux "$arch" $TAG
		else
			echo "use built node linux/$arch/v${NodeVersion}"
		fi
	done
	echo "=============================="
	cd - || exit

	rm -f "$ROOT"/deploy/*.zip
	for arch in "${NODE_ARCHITECTS[@]}"; do
		cp "$ROOT""/../../EdgeNode/dist/edge-node-linux-${arch}-${TAG}-v${NodeVersion}.zip" "$ROOT"/deploy/edge-node-linux-"${arch}"-v"${NodeVersion}".zip
	done

	# build edge-dns
	if [ "$TAG" = "plus" ]; then
		DNS_ROOT=$ROOT"/../../EdgeDNS"
		if [ -d "$DNS_ROOT"  ]; then
			DNSNodeVersion=$(lookup-version "$ROOT""/../../EdgeDNS/internal/const/const.go")
			echo "building edge-dns ${DNSNodeVersion} ..."
			EDGE_DNS_NODE_BUILD_SCRIPT=$ROOT"/../../EdgeDNS/build/build.sh"
			if [ ! -f "$EDGE_DNS_NODE_BUILD_SCRIPT" ]; then
				echo "unable to find edge-dns build script 'EdgeDNS/build/build.sh'"
				exit
			fi
			cd "$ROOT""/../../EdgeDNS/build" || exit
			echo "=============================="
			architects=("amd64" "arm64")
			for arch in "${architects[@]}"; do
				./build.sh linux "$arch" $TAG
			done
			echo "=============================="
			cd - || exit

			for arch in "${architects[@]}"; do
				cp "$ROOT""/../../EdgeDNS/dist/edge-dns-linux-${arch}-v${DNSNodeVersion}.zip" "$ROOT"/deploy/edge-dns-linux-"${arch}"-v"${DNSNodeVersion}".zip
			done
		fi
	fi

	# build sql
	if [ $TAG = "plus" ]; then
		echo "building sql ..."
		"${ROOT}"/sql.sh
	fi

	# copy files
	echo "copying ..."
	if [ ! -d "$DIST" ]; then
		mkdir "$DIST"
		mkdir "$DIST"/bin
		mkdir "$DIST"/configs
		mkdir "$DIST"/logs
		mkdir "$DIST"/data
	fi
	cp "$ROOT"/configs/api.template.yaml "$DIST"/configs/
	cp "$ROOT"/configs/db.template.yaml "$DIST"/configs/
	cp -R "$ROOT"/deploy "$DIST/"
	rm -f "$DIST"/deploy/.gitignore
	cp -R "$ROOT"/installers "$DIST"/

	# building edge installer
	echo "building node installer ..."
	architects=("amd64" "arm64")
	for arch in "${architects[@]}"; do
		# TODO support arm, mips ...
		env GOOS=linux GOARCH="${arch}" go build -trimpath -tags $TAG --ldflags="-s -w" -o "$ROOT"/installers/edge-installer-helper-linux-"${arch}" "$ROOT"/../cmd/installer-helper/main.go
	done

	# building edge dns installer
	if [ $TAG = "plus" ]; then
		echo "building dns node installer ..."
		architects=("amd64" "arm64")
		for arch in "${architects[@]}"; do
			# TODO support arm, mips ...
			env GOOS=linux GOARCH="${arch}" go build -trimpath -tags $TAG --ldflags="-s -w" -o "$ROOT"/installers/edge-installer-dns-helper-linux-"${arch}" "$ROOT"/../cmd/installer-dns-helper/main.go
		done
	fi

	# building api node
	env GOOS="$OS" GOARCH="$ARCH" go build -trimpath -tags $TAG --ldflags="-s -w" -o "$DIST"/bin/edge-api "$ROOT"/../cmd/edge-api/main.go

	# delete hidden files
	find "$DIST" -name ".DS_Store" -delete
	find "$DIST" -name ".gitignore" -delete

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

function lookup-version() {
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
