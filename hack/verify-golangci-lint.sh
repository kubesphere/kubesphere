#!/usr/bin/env bash

# Copyright 2021 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
  echo 'installing golangci-lint '
  pushd "${KUBE_ROOT}/hack/tools" >/dev/null
    GO111MODULE=auto go install -mod= github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2
  popd >/dev/null
fi

cd "${KUBE_ROOT}"

function error_exit {
  if [ $? -eq 1 ]; then
    echo "Please run the following command:"
    echo "  make golint"
  fi
}
trap "error_exit" EXIT

echo 'running golangci-lint '
golangci-lint run \
  --timeout 30m \
  --disable-all \
  -E deadcode \
  -E unused \
  -E varcheck \
  -E ineffassign \
  -E staticcheck \
  -E gosimple \
  -E bodyclose \
  --skip-dirs pkg/client \
  pkg/... cmd/... tools/... test/... kube/...
