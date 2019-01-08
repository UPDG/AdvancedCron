#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
OD="$(pwd)"
# Pushes application version into the build information.
ACRON_VERSION=1.0.0

build(){
	echo Packaging $1 Build
	bdir=acron-${ACRON_VERSION}-$2-$3
	rm -rf builds/$bdir && mkdir -p builds/$bdir
	GOOS=$2 GOARCH=$3 ./build.sh

	mv acron builds/$bdir

	cp README.md builds/$bdir
	cp LICENSE builds/$bdir
	cd builds

	if [ "$2" == "linux" ]; then
		tar -zcf $bdir.tar.gz $bdir
	else
		zip -r -q $bdir.zip $bdir
	fi

	rm -rf $bdir
	cd ..
}

if [ "$1" == "all" ]; then
	rm -rf builds/
	build "Mac" "darwin" "amd64"
	build "Linux" "linux" "amd64"
	build "FreeBSD" "freebsd" "amd64"
	exit
fi

CGO_ENABLED=0 go build -o "$OD/acron" .
