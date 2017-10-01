FROM golang:1.9-alpine


RUN apk --update add musl-dev gcc tar git bash wget && rm -rf /var/cache/apk/*

# Create user
ARG uid=1000
ARG gid=1000
RUN addgroup -g $gid ecs-exporter
RUN adduser -D -u $uid -G ecs-exporter ecs-exporter

RUN mkdir -p /go/src/github.com/slok/ecs-exporter/
RUN chown -R ecs-exporter:ecs-exporter /go

WORKDIR /go/src/github.com/slok/ecs-exporter/

USER ecs-exporter
