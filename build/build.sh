#!/bin/bash

set -uex

# go src
if [ -z "${GOPATH:-}" ]; then
    [ -d go ] || mkdir go

    export GOPATH=$PWD/go
fi

if [ -z "${SRC:-}" ]; then
    SRC=$GOPATH/src/github.com/qmsk/e2

    [ -d $SRC ] || go get -v -d github.com/qmsk/e2/cmd/...
else
    mkdir -p $GOPATH/src/github.com/qmsk
    ln -s $SRC $GOPATH/src/github.com/qmsk/e2
    
    go get -v -u -d github.com/qmsk/e2/cmd/...
fi

GIT_VERSION=$(git -C $SRC describe --tags)

# dist
PACKAGE=qmsk-e2
VERSION=${GIT_VERSION#v}
CMD=(client server tally)

# build static
( cd $SRC/static

    export CI=true

    bower install
)

# prepare base dist
DIST=dist/${PACKAGE}_${VERSION}

tar -C $SRC --exclude-vcs -czvf ${DIST}_src.tar.gz .

install -d $DIST

install -d $DIST/bin
install -m 0755 -t $DIST/bin $SRC/cmd/*/*.sh

install -d $DIST/etc/systemd/system
install -m 0644 -t $DIST/etc/systemd/system $SRC/cmd/*/*.service

rsync -rlpt $SRC/static $DIST/

build_arch () {
    local arch=$1

    # build dist
    DIST_ARCH=${DIST}_$arch

    cp -a $DIST $DIST_ARCH
    install -d $DIST_ARCH/bin

    for cmd in "${CMD[@]}"; do
        go build -o $DIST_ARCH/bin/$cmd -v github.com/qmsk/e2/cmd/$cmd
    done

    tar -czvf $DIST_ARCH.tar.gz $DIST_ARCH/
}

GOOS=linux GOARCH=amd64 build_arch linux-amd64
GOOS=linux GOARCH=arm build_arch linux-arm

