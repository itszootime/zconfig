FROM golang:1.6.2-alpine
MAINTAINER Richard Jones <itszootime@gmail.com>

RUN apk add --update git && rm -rf /var/cache/apk/*
RUN go get -u github.com/tools/godep

WORKDIR /go/src/github.com/itszootime/zconfig

ADD Godeps Godeps
RUN godep restore

ADD . ./
RUN go install

WORKDIR /etc/zconfig
VOLUME ["/etc/zconfig"]
ENTRYPOINT ["zconfig"]
