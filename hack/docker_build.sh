#!/usr/bin/env bash

set -ex
set -o pipefail

# push to kubespheredev with default latest tag
REPO=${REPO:-kubespheredev}
TAG=${TRAVIS_BRANCH:-latest}

# check if build was triggered by a travis cronjob
if [[ -z "$TRAVIS_EVENT_TYPE" ]]; then
    echo "TRAVIS_EVENT_TYPE is empty, also normaly build"
elif [[ $TRAVIS_EVENT_TYPE == "cron" ]]; then
    TAG=dev-$(date +%Y%m%d)
fi


if [[ -z "$DOCKER_PASSWORD" ]]
then
    echo "DOCKER_PASSWORD is empty, also normaly build"

    docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
    docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .

else
    echo "DOCKER_PASSWORD is set, build multi-arch image and push to dockerhub"

    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

    export DOCKER_CLI_EXPERIMENTAL=enabled
    docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
    docker buildx create --name ks-cross-build
    docker buildx use ks-cross-build
    docker buildx inspect --bootstrap
    docker buildx ls

    docker buildx build --platform linux/amd64,linux/arm64 --progress=plain --push -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
    docker buildx build --platform linux/amd64,linux/arm64 --progress=plain --push -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .

    # Clean up
    docker buildx rm ks-cross-build
fi


