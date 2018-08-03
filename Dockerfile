# go backend
FROM golang:1.10.3 as go-build

RUN curl -L -o /tmp/dep-linux-amd64 https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && install -m 0755 /tmp/dep-linux-amd64 /usr/local/bin/dep

WORKDIR /go/src/github.com/qmsk/e2

COPY Gopkg.* ./
RUN dep ensure -vendor-only

COPY . ./
RUN go install -v ./cmd/...


# web frontend
FROM node:9.8.0 as web-build

WORKDIR /go/src/github.com/qmsk/e2/static

COPY static/package.json ./
RUN npm install

COPY static ./


# runtime
# must match with go-build base image
FROM debian:stretch

RUN adduser --system --home /home/qmsk-e2 --uid 1000 --gid 100 qmsk-e2

RUN mkdir -p \
  /opt/qmsk-e2 \
  /opt/qmsk-e2/bin

COPY --from=go-build /go/bin/client /go/bin/server /go/bin/tally /opt/qmsk-e2/bin/
COPY --from=web-build /go/src/github.com/qmsk/e2/static/ /opt/qmsk-e2/static

USER qmsk-e2
ENV PATH=$PATH:/opt/qmsk-e2/bin
CMD ["/opt/qmsk-e2/bin/tally", \
  "--http-listen=:8001", "--http-static=/opt/qmsk-e2/static" \
]
