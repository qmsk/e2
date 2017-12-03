Docker dist builds.

First build the build image, which includes all the build dependencies:

    docker build -t qmsk-e2-build build/

Prepare a directory for the output files, that the build user within the Docker container is able to write to:

    mkdir dist && chmod 0777 dist

Then build the dist packages, from github master:

    docker run --rm -v $PWD/dist:/dist -e DIST=/dist qmsk-e2-build

Alternatively, build from local sources:

    docker run --rm -v $PWD/dist:/dist -e DIST=/dist -v $PWD:/src -e SRC=/src qmsk-e2-build

This will produce the following files:

```
qmsk-e2_0.2.2/
qmsk-e2_0.2.2_linux-amd64/
qmsk-e2_0.2.2_linux-amd64.tar.gz
qmsk-e2_0.2.2_linux-arm/
qmsk-e2_0.2.2_linux-arm.tar.gz
qmsk-e2_0.2.2_src.tar.gz
```
