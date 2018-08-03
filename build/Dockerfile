FROM ubuntu:bionic

RUN apt-get update && apt-get install -y \
    git curl rsync \
    golang-go \
    nodejs npm

RUN curl -L -o /tmp/dep-linux-amd64 https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && install -m 0755 /tmp/dep-linux-amd64 /usr/local/bin/dep

RUN adduser --system --home /home/build --uid 1000 --gid 100 build

USER build
ADD entrypoint.sh build.sh /home/build/

VOLUME /home/build
WORKDIR /home/build
CMD install -d /home/build/go
ENV GOPATH=/home/build/go

ENTRYPOINT ["/home/build/entrypoint.sh"]
CMD ["/home/build/build.sh"]
