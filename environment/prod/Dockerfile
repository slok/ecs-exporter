FROM slok/ecs-exporter_base:latest

USER root

# Prepare
WORKDIR /go/src/github.com/slok/ecs-exporter/
RUN mkdir -p /bin
COPY . ./

# Build
RUN ./build.sh /bin/ecs-exporter
RUN chmod 755 /bin/ecs-exporter

# Clean up
WORKDIR /
RUN rm -rf /go/src/*


EXPOSE 9222

ENTRYPOINT [ "/bin/ecs-exporter" ]
CMD        [ "--help"]
