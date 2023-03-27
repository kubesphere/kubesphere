#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

if ! command -v goimports ; then
# Install goimports
  echo 'installing goimports'
  pushd "${KUBE_ROOT}/hack/tools" >/dev/null
    GO111MODULE=auto go install -mod=mod golang.org/x/tools/cmd/goimports@v0.7.0
  popd >/dev/null
fi

cd "${KUBE_ROOT}" || exit 1

IFS=$'\n' read -r -d '' -a files < <( find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./pkg/apis/*" -not -path "./pkg/client/*" -not -name "zz_generated.deepcopy.go" && printf '\0' )

output=$(goimports -local kubesphere.io/kubesphere -l "${files[@]}")

if [ "${output}" != "" ]; then
    echo "The following files are not import formatted"
    printf '%s\n' "${output[@]}"
    echo "Please run the following command:"
    echo "make goimports"
    exit 1
fi
