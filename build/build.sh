#!/bin/bash

PACKAGE=qmsk-e2
CMD=(client server tally)

set -uex

echo "building package=$PACKAGE version=$VERSION at $DIST"

# prepare base/src dist
dist=${PACKAGE}_${VERSION}

tar --exclude-vcs --exclude=*.tar.gz --exclude=dist -czvf $DIST/${dist}_src.tar.gz .

install -d $DIST/$dist

install -d $DIST/$dist/bin
install -m 0755 -t $DIST/$dist/bin ./cmd/*/*.sh

rsync -ax ./static $DIST/$dist/

build_arch () {
    local arch=$1

    # build dist
    dist_arch=${dist}_$arch

    rsync -ax $DIST/$dist/ $DIST/$dist_arch
    install -d $DIST/$dist_arch/bin

    for cmd in "${CMD[@]}"; do
        go build -o $DIST/$dist_arch/bin/$cmd -v github.com/qmsk/e2/cmd/$cmd
    done

    tar -C $DIST -czvf $DIST/$dist_arch.tar.gz $dist_arch/

    # debian package
    install -d $DIST/$dist_arch.pkg
    install -d $DIST/$dist_arch.pkg/opt/qmsk-e2

    rsync -ax build/DEBIAN $DIST/$dist_arch.pkg
    rsync -ax $DIST/$dist_arch/{bin,static} $DIST/$dist_arch.pkg/opt/qmsk-e2/

    install -d $DIST/$dist_arch.pkg/lib/systemd/system
    install -d $DIST/$dist_arch.pkg/etc/default

    install -m 0644 -t $DIST/$dist_arch.pkg/lib/systemd/system ./build/systemd/*.service
    install -m 0644 -t $DIST/$dist_arch.pkg/etc/default/ ./build/etc/default/*

    sed -i "s/@VERSION@/$VERSION/" $DIST/$dist_arch.pkg/DEBIAN/*
    sed -i "s/@ARCH@/$DEB_ARCH/" $DIST/$dist_arch.pkg/DEBIAN/*

    dpkg-deb -b $DIST/$dist_arch.pkg $DIST
}

GOOS=linux GOARCH=amd64 DEB_ARCH=amd64 build_arch linux-amd64
GOOS=linux GOARCH=arm DEB_ARCH=armhf build_arch linux-arm

ls -1p $DIST/
sha256sum $DIST/${dist}_*.tar.gz $DIST/${dist}_*.deb > $DIST/SHA256SUM
