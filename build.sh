#!/usr/bin/env bash

docker build -t storefinder/cli . -f Dockerfile && \
docker create --name storefinder-cli storefinder/cli && \
docker cp storefinder-cli:./storefinder . && \
docker rm -f storefinder-cli
