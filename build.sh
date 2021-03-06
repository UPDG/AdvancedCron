#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
OD="$(pwd)"
# Pushes application version into the build information.
ACRON_VERSION=1.1.0

build(){
	echo Packaging $1 Build
	bdir=acron-${ACRON_VERSION}-$2-$3
	rm -rf builds/$bdir && mkdir -p builds/$bdir

	cd build
	docker-compose run builder ./build/build_platform.sh $2/$3
	cd ..

    mv build/acron-$2-$3 build/acron
	mv build/acron builds/$bdir

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
	exit
fi

CGO_ENABLED=1 go build -o "$OD/acron" .
