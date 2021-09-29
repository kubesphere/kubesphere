#!/bin/bash

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

KIND_LOG_LEVEL="1"

if [ -n "${DEBUG}" ]; then
  set -x
  KIND_LOG_LEVEL="6"
fi

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}"/hack/lib/init.sh

cleanup() {
  kind delete cluster \
    --verbosity="${KIND_LOG_LEVEL}" \
    --name "${KIND_CLUSTER_NAME}"
}

trap cleanup EXIT

export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-kubesphere-e2e}

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

echo "Creating Kubernetes cluster with kind"

export K8S_VERSION=${K8S_VERSION:-v1.21.1}

kind create cluster \
  --verbosity="${KIND_LOG_LEVEL}" \
  --name "${KIND_CLUSTER_NAME}" \
  --config "${KUBE_ROOT}"/test/e2e/kind.yaml \
  --retain \
  --image kindest/node:"${K8S_VERSION}"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "Deploy KubeSphere"
"${KUBE_ROOT}"/hack/deploy-kubesphere.sh

echo "Run e2e test"
go test ./test/e2e

