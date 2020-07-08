#!/usr/bin/env bash

set -ex
set -o pipefail

tag_for_branch() {
    local tag=$1
    if [[ "${tag}" == "master" ]]; then
        tag="latest"
    fi
    echo ${tag}
}

# push to kubespheredev with default latest tag
REPO=${REPO:-kubespheredev}
TAG=$(tag_for_branch $1)

docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .

# Push image to dockerhub, need to support multiple push
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push $REPO/ks-apiserver:$TAG
docker push $REPO/ks-controller-manager:$TAG
