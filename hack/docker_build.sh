#!/usr/bin/env bash

set -ex
set -o pipefail

# Default image repo
REPO=kubespheredev
# Set tag to latest if no argument was given, normally was branch name
TAG=${1:-latest}

docker build -f build/ks-apigateway/Dockerfile -t $REPO/ks-apigateway:$TAG .
docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
docker build -f build/ks-iam/Dockerfile -t $REPO/ks-account:$TAG .
docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .
docker build -f build/hypersphere/Dockerfile -t $REPO/hypersphere:$TAG .
docker build -f ./pkg/db/Dockerfile -t $REPO/ks-devops:flyway-$TAG ./pkg/db/

# Push image to dockerhub, need to support multiple push

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push $REPO/ks-apigateway:$TAG
docker push $REPO/ks-apiserver:$TAG
docker push $REPO/ks-account:$TAG
docker push $REPO/ks-controller-manager:$TAG
docker push $REPO/hypersphere:$TAG
docker push $REPO/ks-devops:flyway-$TAG
