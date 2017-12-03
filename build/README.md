Docker dist builds.

First build the build image, which includes all the build dependencies:

    docker build -t qmsk-e2-build build/

Then build the dist packages, from github master:

    docker run --rm -v $PWD:/home/build qmsk-e2-build

Alternatively, build from local sources:

    docker run --rm -v $PWD:/home/build -v $PWD:/home/build/src -e SRC=/home/build/src qmsk-e2-build

This will produce the following files:

```
qmsk-e2_0.2.2/
qmsk-e2_0.2.2_linux-amd64/
qmsk-e2_0.2.2_linux-amd64.tar.gz
qmsk-e2_0.2.2_linux-arm/
qmsk-e2_0.2.2_linux-arm.tar.gz
qmsk-e2_0.2.2_src.tar.gz
```
