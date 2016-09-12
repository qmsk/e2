#!/bin/bash

set -uex

# go src
if [ -z "${GOPATH:-}" ]; then
    mkdir go

    export GOPATH=$PWD/go
fi

if [ -z "${SRC:-}" ]; then
    SRC=$GOPATH/src/github.com/qmsk/e2

    go get -v -d github.com/qmsk/e2/cmd/...
fi

GIT_VERSION=$(git -C $SRC describe --tags)

# dist
PACKAGE=qmsk-e2
VERSION=${GIT_VERSION#v}
DIST=dist/${PACKAGE}_${VERSION}

# build static
( cd $SRC/static

    export CI=true

    bower install
)

# build dist
install -d $DIST

install -d $DIST/bin
install -m 0755 -t $DIST/bin $SRC/cmd/tally/*.sh

install -d $DIST/etc/systemd/system
install -m 0644 -t $DIST/etc/systemd/system $SRC/dist/etc/systemd/system/*.service

rsync -rlpt $SRC/static $DIST/

# build arch
go get github.com/qmsk/e2/cmd/...

install -d $DIST/bin
install -m 0755 -t $DIST/bin $GOPATH/bin/client
install -m 0755 -t $DIST/bin $GOPATH/bin/server
install -m 0755 -t $DIST/bin $GOPATH/bin/tally

tar -czvf $DIST.tar.gz $DIST/
