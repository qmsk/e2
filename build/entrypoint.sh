#!/bin/bash

set -uex

# go src
if [ -z "${GOPATH:-}" ]; then
    export GOPATH=$HOME/go
fi

[ -d $GOPATH ] || mkdir $GOPATH

if [ -z "${SRC:-}" ]; then
    SRC=$GOPATH/src/github.com/qmsk/e2

    [ -d $SRC ] || go get -v -d github.com/qmsk/e2/...
else
    mkdir -p $GOPATH/src/github.com/qmsk
    ln -s $SRC $GOPATH/src/github.com/qmsk/e2

    go get -v -u -d github.com/qmsk/e2/...
fi

[ -z "${GIT_TAG:-}" ] || git -C $SRC checkout $GIT_TAG

GIT_VERSION=$(git -C $SRC describe --tags)

export DIST=${DIST:-$PWD}
export VERSION=${VERSION:-${GIT_VERSION#v}}

cd $SRC && exec "$@"
