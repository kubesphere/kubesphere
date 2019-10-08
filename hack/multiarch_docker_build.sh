#!/usr/bin/env bash

set -ex
set -o pipefail

export DOCKER_CLI_EXPERIMENTAL=enabled
REPO=kubespheredev
TAG=latest

uname -m
sudo docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker run --rm -t arm64v8/ubuntu uname -m

sudo docker buildx create --name ks-all
sudo docker buildx use ks-all
sudo docker buildx inspect --bootstrap
sudo docker buildx ls


#echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
sudo docker buildx build -f  build/ks-apigateway/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 --progress=plain -t $REPO/ks-apigateway:$TAG .
sudo docker buildx build -f  build/ks-apiserver/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 --progress=plain -t $REPO/ks-apiserver:$TAG .
sudo docker buildx build -f  build/ks-iam/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 --progress=plain -t $REPO/ks-iam:$TAG .
sudo docker buildx build -f  build/ks-controller-manager/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 --progress=plain -t $REPO/ks-controller-manager:$TAG .
sudo docker buildx build -f  build/hypersphere/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 --progress=plain -t $REPO/hypersphere:$TAG .
sudo docker buildx build -f ./pkg/db/Dockerfile --output=type=registry --platform linux/amd64,linux/arm64 -t $REPO/ks-devops:flyway-$TAG --progress=plain ./pkg/db/
