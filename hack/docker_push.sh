#!/usr/bin/env bash

# Push image to dockerhub, need to support multiple push

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push kubespheredev/ks-apiserver:latest