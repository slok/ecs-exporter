#!/bin/bash -x

if [ -n "$TRAVIS_PULL_REQUEST_BRANCH" ]; then
  # running PR pipeline
  VERSION="feature-branch-$TRAVIS_PULL_REQUEST_BRANCH-$TRAVIS_COMMIT"
elif [ -n "$TRAVIS_TAG" ]; then
  # running git tag pipeline
  VERSION=$TRAVIS_TAG
fi

go build -o "${1:-./bin/ecs-exporter}" --ldflags "-w -linkmode external -extldflags '-static'" "./cmd/ecs-exporter/"
