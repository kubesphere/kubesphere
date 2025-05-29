#!/usr/bin/env bash

# This script checks coding style for go language files in each
# Kubernetes package by golint.
# Usage: `hack/verify-golangci-lint.sh`.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

kube::golang::setup_env

if ! command -v golangci-lint ; then
  # Install golangci-lint
  echo 'installing golangci-lint'
  go install -mod=mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
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
  --default standard \
  --new-from-rev=HEAD^ \
  pkg/... \
  cmd/... \
  staging/... \
  tools/...
