Docker dist builds.

First build the build image, which includes all the build dependencies:

    docker build -t qmsk-e2-build build/ 

Then build the dist package, from github master:

    docker run -v $PWD/dist:/home/build/dist qmsk-e2-build

Alternatively, build from local sources:

    docker run -v $PWD/dist:/home/build/dist -v $PWD:/home/build/src -e SRC=/home/build/src qmsk-e2-build
