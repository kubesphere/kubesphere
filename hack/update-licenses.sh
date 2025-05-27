#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

kube::golang::setup_env

if ! command -v goimports ; then
  # Install goimports
  echo 'installing goimports'
  go install -mod=mod github.com/apache/skywalking-eyes/cmd/license-eye@v0.7.0
fi

cd "${KUBE_ROOT}" || exit 1

echo 'running skywalking-eyes fix '

mapfile -t files < <(git diff --name-only HEAD~1 | grep -v '^vendor/' | grep -E '\.go$' || true)

if [ "${#files[@]}" -eq 0 ]; then
  echo "✅ No files changed."
  exit 0
fi

license-eye header fix "${files[@]}"