#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

if ! command -v license-eye &> /dev/null
then
    # Ensure that we find the binaries we build before anything else.
    export GOBIN="${KUBE_OUTPUT_BINPATH}"
    PATH="${GOBIN}:${PATH}"

    # Explicitly opt into go modules, even though we're inside a GOPATH directory
    export GO111MODULE=on
    # Explicitly clear GOFLAGS, since GOFLAGS=-mod=vendor breaks dependency resolution while rebuilding vendor
    export GOFLAGS=

    # Install skywalking-eyes
    echo 'installing skywalking-eyes '
    go install -mod=mod github.com/apache/skywalking-eyes/cmd/license-eye@v0.4.0
fi

cd "${KUBE_ROOT}"

echo 'running skywalking-eyes fix '
license-eye header fix
exit 0
