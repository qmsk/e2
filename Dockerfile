FROM ubuntu:xenial

RUN apt-get update && apt-get install -y \
    git \
    golang-go \
    nodejs nodejs-legacy npm

RUN npm install -g bower
ENV GOPATH=/go

RUN adduser --system --home /home/build --uid 1000 --gid 100 qmsk-e2
RUN install -o qmsk-e2 -d /go/src/github.com/qmsk/e2/static

ADD static/bower.json /go/src/github.com/qmsk/e2/static/
WORKDIR /go/src/github.com/qmsk/e2/static/
USER qmsk-e2
RUN CI=true bower install

ADD . /go/src/github.com/qmsk/e2
WORKDIR /go/src/github.com/qmsk/e2
USER root
RUN go get ./cmd/...

USER qmsk-e2
ENV PATH=/go/bin:$PATH
