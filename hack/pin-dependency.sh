#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

# Usage:
#   hack/pin-dependency.sh $MODULE $SHA-OR-TAG
#
# Example:
#   hack/pin-dependency.sh github.com/docker/docker 501cb131a7b7

# Explicitly opt into go modules, even though we're inside a GOPATH directory
export GO111MODULE=on
# Explicitly clear GOFLAGS, since GOFLAGS=-mod=vendor breaks dependency resolution while rebuilding vendor
export GOFLAGS=
# Detect problematic GOPROXY settings that prevent lookup of dependencies
if [[ "${GOPROXY:-}" == "off" ]]; then
  kube::log::error "Cannot run with \$GOPROXY=off"
  exit 1
fi

kube::golang::verify_go_version
kube::util::require-jq

dep="${1:-}"
sha="${2:-}"
if [[ -z "${dep}" || -z "${sha}" ]]; then
  echo "Usage:"
  echo "  hack/pin-dependency.sh \$MODULE \$SHA-OR-TAG"
  echo ""
  echo "Example:"
  echo "  hack/pin-dependency.sh github.com/docker/docker 501cb131a7b7"
  echo ""
  exit 1
fi

_tmp="${KUBE_ROOT}/_tmp"
cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT
cleanup
mkdir -p "${_tmp}"

# Add the require directive
echo "Running: go get ${dep}@${sha}"
go get -d "${dep}@${sha}"

# Find the resolved version
rev=$(go mod edit -json | jq -r ".Require[] | select(.Path == \"${dep}\") | .Version")

# No entry in go.mod, we must be using the natural version indirectly
if [[ -z "${rev}" ]]; then
  # backup the go.mod file, since go list modifies it
  cp go.mod "${_tmp}/go.mod.bak"
  # find the revision
  rev=$(go list -m -json "${dep}" | jq -r .Version)
  # restore the go.mod file
  mv "${_tmp}/go.mod.bak" go.mod
fi

# No entry found
if [[ -z "${rev}" ]]; then
  echo "Could not resolve ${sha}"
  exit 1
fi

echo "Resolved to ${dep}@${rev}"

# Add the replace directive
echo "Running: go mod edit -replace ${dep}=${dep}@${rev}"
go mod edit -replace "${dep}=${dep}@${rev}"

echo ""
echo "Run hack/update-vendor.sh to rebuild the vendor directory"
