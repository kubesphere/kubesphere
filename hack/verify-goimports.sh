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

. "$(dirname "${BASH_SOURCE[0]}")/lib/init.sh"

cd "${KUBE_ROOT}/hack" || exit 1

if ! command -v goimports &> /dev/null
then
    echo "goimports could not be found on your machine, please install it first"
    exit 1
fi

cd "${KUBE_ROOT}" || exit 1

IFS=$'\n' read -r -d '' -a files < <( find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./pkg/apis/*" -not -path "./pkg/client/*" && printf '\0' )

output=$(goimports -local kubesphere.io/kubesphere -l "${files[@]}")

if [ "${output}" != "" ]; then
    echo "The following files are not import formatted "
    printf '%s\n' "${output[@]}"
    echo "Please run the following command:"
    echo "  make goimports"
    exit 1
fi
