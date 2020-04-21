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


docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .

# Push image to dockerhub, need to support multiple push

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push $REPO/ks-apiserver:$TAG
docker push $REPO/ks-controller-manager:$TAG
