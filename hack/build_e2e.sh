#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

VERBOSE=${VERBOSE:-"0"}
# V=""
if [[ "${VERBOSE}" == "1" ]];then
    # V="-x"
    set -x
fi

# ROOTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OUTPUT_DIR=bin
BUILDPATH=./${1:?"path to build"}
OUT=${OUTPUT_DIR}/${1:?"output path"}

BUILD_GOOS=${GOOS:-$(go env GOOS)}
BUILD_GOARCH=${GOARCH:-$(go env GOARCH)}
GOBINARY=${GOBINARY:-go}
LDFLAGS=$(kube::version::ldflags)

time GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} ${GOBINARY} test \
        -c \
        -ldflags "${LDFLAGS}" \
        -o "${OUT}" \
        "${BUILDPATH}"
