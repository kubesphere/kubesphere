#!/usr/bin/env bash

set -ex
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

# push to kubesphere with default latest tag
TAG=${TAG:-latest}
REPO=${REPO:-kubesphere}
PUSH=${PUSH:-}

# support other container tools. e.g. podman
CONTAINER_CLI=${CONTAINER_CLI:-docker}
CONTAINER_BUILDER=${CONTAINER_BUILDER:-"buildx build"}

# If set, just building, no pushing
if [[ -z "${DRY_RUN:-}" ]]; then
  PUSH="--push"
fi

# supported platforms
PLATFORMS=linux/amd64,linux/arm64

${CONTAINER_CLI} ${CONTAINER_BUILDER} \
  --platform ${PLATFORMS} \
  ${PUSH} \
  -f build/ks-apiserver/Dockerfile \
  -t "${REPO}"/ks-apiserver:"${TAG}" .

${CONTAINER_CLI} ${CONTAINER_BUILDER} \
  --platform ${PLATFORMS} \
  ${PUSH} \
  -f build/ks-controller-manager/Dockerfile \
  -t "${REPO}"/ks-apiserver:"${TAG}" .

