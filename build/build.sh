#!/bin/bash

PACKAGE=qmsk-e2
DIST=${DIST:-.}
CMD=(client server tally)

set -uex

echo "building package=$PACKAGE version=$VERSION at $DIST"

# build static
( cd ./static

    export CI=true

    bower install
)

# prepare base/src dist
dist=${PACKAGE}_${VERSION}

tar --exclude-vcs -czvf $DIST/${dist}_src.tar.gz .

install -d $DIST/$dist

install -d $DIST/$dist/bin
install -m 0755 -t $DIST/$dist/bin ./cmd/*/*.sh

install -d $DIST/$dist/etc/systemd/system
install -m 0644 -t $DIST/$dist/etc/systemd/system ./cmd/*/*.service

rsync -rlpt ./static $DIST/$dist/

build_arch () {
    local arch=$1

    # build dist
    dist_arch=${dist}_$arch

    cp -a $DIST/$dist $DIST/$dist_arch
    install -d $DIST/$dist_arch/bin

    for cmd in "${CMD[@]}"; do
        go build -o $DIST/$dist_arch/bin/$cmd -v github.com/qmsk/e2/cmd/$cmd
    done

    tar -czvf $DIST/$dist_arch.tar.gz $DIST/$dist_arch/
}

GOOS=linux GOARCH=amd64 build_arch linux-amd64
GOOS=linux GOARCH=arm build_arch linux-arm
