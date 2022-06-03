#!/usr/bin/env bash
# Copyright 2022 The KubeSphere Authors.
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
#

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
    pushd "${KUBE_ROOT}/hack/tools" >/dev/null
      go install github.com/apache/skywalking-eyes/cmd/license-eye@v0.2.0
    popd >/dev/null
fi

cd "${KUBE_ROOT}"

echo 'running skywalking-eyes check '
license-eye header check
exit 0
