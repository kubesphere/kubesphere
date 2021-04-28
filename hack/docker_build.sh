#!/usr/bin/env bash

set -ex
set -o pipefail

tag_for_branch() {
    local tag=$1
    if [[ "${tag}" == "" ]]; then
        tag=$(git branch --show-current)
        tag=${tag/\//-}
    fi

    if [[ "${tag}" == "master" ]]; then
        tag="latest"
    fi
    echo ${tag}
}

get_repo() {
    local repo=${REPO} # read from env
    repo=${repo:-kubespheredev}
    if [[ "$1" != "" ]]; then
      repo="$1"
    fi

    # set the default value if there's no user defined
    if [[ "${repo}" == "" ]]; then
      repo="kubespheredev"
    fi
    echo "$repo"
}

# push to kubespheredev with default latest tag
TAG=$(tag_for_branch "$1")
REPO=$(get_repo "$2")

# Push image to dockerhub, need to support multiple push
cat ~/.docker/config.json | grep index.docker.io
if [[ $? != 0 ]]; then
  echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
fi

docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
docker push $REPO/ks-apiserver:$TAG
# print the full docker image path for your convience
docker images --digests | grep $REPO/ks-apiserver | grep $TAG | awk '{print $1":"$2"@"$3}'

docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .
docker push $REPO/ks-controller-manager:$TAG
# print the full docker image path for your convience
docker images --digests | grep $REPO/ks-controller-manager | grep $TAG | awk '{print $1":"$2"@"$3}'
