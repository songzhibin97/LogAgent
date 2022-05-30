FROM golang:alpine as builder

WORKDIR /binbin
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

LABEL MAINTAINER="binbin@qq.com"

WORKDIR /binbin

COPY --from=0 /binbin/server ./
COPY --from=0 /binbin/config.yaml ./

ENTRYPOINT ./server -c config.yaml
