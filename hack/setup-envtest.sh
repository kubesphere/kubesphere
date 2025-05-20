#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

kube::golang::setup_env

if ! command -v setup-envtest ; then
  echo 'installing setup-envtest'
  # While it's preferable not to use @latest here, we have no choice at the moment. Details at
  # https://github.com/kubernetes-sigs/kubebuilder/issues/2480
  go install -mod=mod sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20240521074430-fbb7d370bebc
fi
