#!/usr/bin/env bash

# This script is modified version of Kubernetes script

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=auto

# The root of the build/dist directory
KUBE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

KUBE_OUTPUT_SUBPATH="${KUBE_OUTPUT_SUBPATH:-_output/local}"
KUBE_OUTPUT="${KUBE_ROOT}/${KUBE_OUTPUT_SUBPATH}"
KUBE_OUTPUT_BINPATH="${KUBE_OUTPUT}/bin"

export THIS_PLATFORM_BIN="${KUBE_ROOT}/_output/bin"

source "${KUBE_ROOT}/hack/lib/util.sh"
source "${KUBE_ROOT}/hack/lib/logging.sh"

kube::log::install_errexit

source "${KUBE_ROOT}/hack/lib/golang.sh"

KUBE_OUTPUT_HOSTBIN="${KUBE_OUTPUT_BINPATH}/$(kube::util::host_platform)"
export KUBE_OUTPUT_HOSTBIN


# This emulates "readlink -f" which is not available on MacOS X.
# Test:
# T=/tmp/$$.$RANDOM
# mkdir $T
# touch $T/file
# mkdir $T/dir
# ln -s $T/file $T/linkfile
# ln -s $T/dir $T/linkdir
# function testone() {
#   X=$(readlink -f $1 2>&1)
#   Y=$(kube::readlinkdashf $1 2>&1)
#   if [ "$X" != "$Y" ]; then
#     echo readlinkdashf $1: expected "$X", got "$Y"
#   fi
# }
# testone /
# testone /tmp
# testone $T
# testone $T/file
# testone $T/dir
# testone $T/linkfile
# testone $T/linkdir
# testone $T/nonexistant
# testone $T/linkdir/file
# testone $T/linkdir/dir
# testone $T/linkdir/linkfile
# testone $T/linkdir/linkdir
function kube::readlinkdashf {
  # run in a subshell for simpler 'cd'
  (
    if [[ -d "${1}" ]]; then # This also catch symlinks to dirs.
      cd "${1}"
      pwd -P
    else
      cd "$(dirname "${1}")"
      local f
      f=$(basename "${1}")
      if [[ -L "${f}" ]]; then
        readlink "${f}"
      else
        echo "$(pwd -P)/${f}"
      fi
    fi
  )
}

# This emulates "realpath" which is not available on MacOS X
# Test:
# T=/tmp/$$.$RANDOM
# mkdir $T
# touch $T/file
# mkdir $T/dir
# ln -s $T/file $T/linkfile
# ln -s $T/dir $T/linkdir
# function testone() {
#   X=$(realpath $1 2>&1)
#   Y=$(kube::realpath $1 2>&1)
#   if [ "$X" != "$Y" ]; then
#     echo realpath $1: expected "$X", got "$Y"
#   fi
# }
# testone /
# testone /tmp
# testone $T
# testone $T/file
# testone $T/dir
# testone $T/linkfile
# testone $T/linkdir
# testone $T/nonexistant
# testone $T/linkdir/file
# testone $T/linkdir/dir
# testone $T/linkdir/linkfile
# testone $T/linkdir/linkdir
kube::realpath() {
  if [[ ! -e "${1}" ]]; then
    echo "${1}: No such file or directory" >&2
    return 1
  fi
  kube::readlinkdashf "${1}"
}