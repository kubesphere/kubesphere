#!/bin/bash
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

if ! command -v setup-envtest ; then
  echo 'installing setup-envtest'
  # While it's preferable not to use @latest here, we have no choice at the moment. Details at
  # https://github.com/kubernetes-sigs/kubebuilder/issues/2480
  GO111MODULE=auto go install -mod=mod sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20240521074430-fbb7d370bebc
fi

setup-envtest use 1.23.x --bin-dir="${KUBE_OUTPUT_BINPATH}"
