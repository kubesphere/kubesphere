#!/usr/bin/env bash

# Copyright 2014 The Kubernetes Authors.
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
source "${KUBE_ROOT}/hack/lib/util.sh"

# Excluded check patterns are always skipped.
EXCLUDED_PATTERNS=(
  "verify-all.sh"                # this script calls the make rule and would cause a loop
  "verify-linkcheck.sh"          # runs in separate Jenkins job once per day due to high network usage
  "verify-*-dockerized.sh"       # Don't run any scripts that intended to be run dockerized
  "verify-govet-levee.sh"        # Do not run levee analysis by default while KEP-1933 implementation is in alpha.
  "verify-licenses.sh"
  "verify-shellcheck.sh"
)

while IFS='' read -r line; do EXCLUDED_CHECKS+=("$line"); done < <(ls "${EXCLUDED_PATTERNS[@]/#/${KUBE_ROOT}/hack/}" 2>/dev/null || true)
TARGET_LIST=()

function is-excluded {
  for e in "${EXCLUDED_CHECKS[@]}"; do
    if [[ $1 -ef "${e}" ]]; then
      return
    fi
  done
  return 1
}

# Collect Failed tests in this Array , initialize it to nil
FAILED_TESTS=()

function print-failed-tests {
  echo -e "========================"
  echo -e "${color_red:?}FAILED TESTS${color_norm:?}"
  echo -e "========================"
  for t in "${FAILED_TESTS[@]}"; do
      echo -e "${color_red}${t}${color_norm}"
  done
}

function run-checks {
  local -r pattern=$1
  local -r runner=$2

  local t
  for t in ${pattern}
  do
    local check_name
    check_name="$(basename "${t}")"

    if is-excluded "${t}" ; then 
      echo "Skipping ${check_name}"
      continue
    fi

    echo -e "Verifying ${check_name}"
    local start
    start=$(date +%s)
    "${runner}" "${t}" && tr=$? || tr=$?
    local elapsed=$(($(date +%s) - start))
    if [[ ${tr} -eq 0 ]]; then
      echo -e "${color_green:?}SUCCESS${color_norm}  ${check_name}\t${elapsed}s"
    else
      echo -e "${color_red}FAILED${color_norm}   ${check_name}\t${elapsed}s"
      ret=1
      FAILED_TESTS+=("${t}")
    fi
  done
}

# Check invalid targets specified in "WHAT" and mark them as failure cases
function missing-target-checks {
  # In case WHAT is not specified
  [[ ${#TARGET_LIST[@]} -eq 0 ]] && return

  for v in "${TARGET_LIST[@]}"
  do
    [[ -z "${v}" ]] && continue

    FAILED_TESTS+=("${v}")
    ret=1
  done
}


ret=0
run-checks "${KUBE_ROOT}/hack/verify-*.sh" bash
missing-target-checks

if [[ ${ret} -eq 1 ]]; then
    print-failed-tests
fi
exit ${ret}
