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

# Detect problematic GOPROXY settings that prevent lookup of dependencies
if [[ "${GOPROXY:-}" == "off" ]]; then
  kube::log::error "Cannot run with \$GOPROXY=off"
  exit 1
fi

kube::golang::setup_env

if ! command -v goimports ; then
  # Install goimports
  echo 'installing goimports'
  go install -mod=mod golang.org/x/tools/cmd/goimports@v0.33.0
fi

cd "${KUBE_ROOT}" || exit 1

IFS=$'\n' read -r -d '' -a files < <( find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./_output/*" -not -name "zz_generated.deepcopy.go" && printf '\0' )

output=$(goimports -local kubesphere.io/kubesphere -l "${files[@]}")

if [ "${output}" != "" ]; then
    echo "The following files are not import formatted"
    printf '%s\n' "${output[@]}"
    echo "Please run the following command:"
    echo "make goimports"
    exit 1
fi
