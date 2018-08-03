#!/bin/bash

set -uex

if [ -z "${SRC:-}" ]; then
    SRC=$GOPATH/src/github.com/qmsk/e2

    [ -d $SRC ] || go get -v -d github.com/qmsk/e2/...
else
    mkdir -p $GOPATH/src/github.com/qmsk
    cp -ar $SRC $GOPATH/src/github.com/qmsk/e2
fi

if [ -n "${GIT_TAG:-}" ]; then
  git -C $SRC checkout $GIT_TAG
fi

GIT_VERSION=$(git -C $SRC describe --tags)

export DIST=${DIST:-$PWD}
export VERSION=${VERSION:-${GIT_VERSION#v}}

cd $GOPATH/src/github.com/qmsk/e2

dep ensure -vendor-only

exec "$@"
