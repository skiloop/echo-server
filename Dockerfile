FROM golang:1.10 AS base

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go get github.com/fatedier/frp

FROM scratch

COPY --from=base /go/bin/echo-server /echo-server

WORKDIR /

EXPOSE 8000

ENTRYPOINT ["/echo-server"]
