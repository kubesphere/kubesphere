#!/usr/bin/env bash

# This script checks coding style for go language files in each
# Kubernetes package by golint.
# Usage: `hack/verify-golangci-lint.sh`.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"
source "${KUBE_ROOT}/hack/lib/util.sh"

kube::golang::verify_go_version

# Ensure that we find the binaries we build before anything else.
export GOBIN="${KUBE_OUTPUT_BINPATH}"
PATH="${GOBIN}:${PATH}"

# Explicitly opt into go modules, even though we're inside a GOPATH directory
export GO111MODULE=on

if ! command -v golangci-lint ; then
# Install golangci-lint
  echo 'installing golangci-lint'
  GO111MODULE=auto go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2
fi

cd "${KUBE_ROOT}"

function error_exit {
  if [ $? -eq 1 ]; then
    echo "Please run the following command:"
    echo "make golint"
  fi
}
trap "error_exit" EXIT

go build ./...

echo "running golangci-lint: REV=HEAD^"
golangci-lint run \
  -v \
  --timeout 30m \
  --disable-all \
  -E unused \
  -E ineffassign \
  -E staticcheck \
  -E gosimple \
  -E bodyclose \
  --skip-dirs pkg/client \
  --new-from-rev=HEAD^ \
  ./...
