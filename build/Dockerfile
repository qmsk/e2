FROM ubuntu:xenial

RUN apt-get update && apt-get install -y \
    git \
    golang-go \
    nodejs nodejs-legacy npm

RUN npm install -g bower

RUN adduser --system --home /home/build --uid 1000 --gid 100 build

USER build
ADD build.sh /home/build/

VOLUME /home/build
WORKDIR /home/build

CMD ["/home/build/build.sh"]
