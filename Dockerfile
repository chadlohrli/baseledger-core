FROM golang:1.16 AS builder

RUN mkdir -p /go/src/github.com/providenetwork
ADD . /go/src/github.com/providenetwork/baseledger-core

WORKDIR /go/src/github.com/providenetwork/baseledger-core
RUN make build

FROM alpine

RUN apk add --no-cache bash

RUN mkdir -p /baseledger
WORKDIR /baseledger

COPY --from=builder /go/src/github.com/providenetwork/baseledger-core/.bin /baseledger/.bin
COPY --from=builder /go/src/github.com/providenetwork/baseledger-core/ops /baseledger/ops

EXPOSE 1337
EXPOSE 33333

ENTRYPOINT ["./.bin/node"]
