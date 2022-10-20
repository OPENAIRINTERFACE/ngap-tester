FROM golang:1.19.1-bullseye AS sim

WORKDIR $GOPATH/src/gnbsim

COPY . $GOPATH/src/ngap-tester 

RUN cd $GOPATH/src/ngap-tester && \
    go build -buildvcs=false -mod=vendor

FROM alpine:3.16 AS ngap-tester
ENV GOPATH=/go

RUN apk update && apk add -U gcompat strace net-tools curl netcat-openbsd bind-tools bash

WORKDIR /ngap-tester/bin

COPY --from=sim $GOPATH/src/ngap-tester/ngap-tester /ngap-tester/bin/

CMD ["/ngap-tester/bin/ngap-tester", "run", "--all"]
