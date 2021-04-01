#!/usr/bin/env bash

set -e
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

# push to $REPO with default $TAG tag
TAG=$(tag_for_branch "$1")
REPO=$(get_repo "$2")

docker build -f build/ks-apiserver/Dockerfile -t $REPO/ks-apiserver:$TAG .
# print the full docker image path for your convience
docker images --digests | grep $REPO/ks-apiserver | grep $TAG | awk '{print $1":"$2"@"$3}'

docker build -f build/ks-controller-manager/Dockerfile -t $REPO/ks-controller-manager:$TAG .
# print the full docker image path for your convience
docker images --digests | grep $REPO/ks-controller-manager | grep $TAG | awk '{print $1":"$2"@"$3}'
