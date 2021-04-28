#!/usr/bin/env bash

set -ex
set -o pipefail
BUILDPLATFORM="linux/amd64,linux/arm64"
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

TAG=$TAG-multiarch
docker buildx build --platform=${BUILDPLATFORM} \
                    -f build/Dockerfile \
                    -t $REPO/ks-apiserver:$TAG . \
                    --target=ks-apiserver --push

docker buildx build --platform=${BUILDPLATFORM} \
                    -f build/Dockerfile \
                    -t $REPO/ks-controller-manager:$TAG . \
                    --target=ks-controller-manager --push