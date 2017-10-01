FROM slok/ecs-exporter_base:latest
USER root
RUN apk add --no-cache g++

USER ecs-exporter

RUN go get github.com/golang/dep/cmd/dep
RUN go get github.com/golang/mock/mockgen
