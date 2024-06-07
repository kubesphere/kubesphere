#!/bin/bash

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
