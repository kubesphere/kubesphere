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

# A single script that runs a predefined set of update-* scripts, as they often go together.
set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/../..
source "${KUBE_ROOT}/hack/lib/init.sh"

# If called directly, exit.
if [[ "${CALLED_FROM_MAIN_MAKEFILE:-""}" == "" ]]; then
    echo "ERROR: $0 should not be run directly." >&2
    echo >&2
    echo "Please run this command using \"make update\""
    exit 1
fi

SILENT=${SILENT:-true}
ALL=${FORCE_ALL:-false}
V=""
if [[ "${SILENT}" != "true" ]]; then
	V="-v"
fi

trap 'exit 1' SIGINT

if ${SILENT} ; then
	echo "Running in silent mode, run with SILENT=false if you want to see script logs."
fi

if ! ${ALL} ; then
	echo "Running in short-circuit mode; run with FORCE_ALL=true to force all scripts to run."
fi

"${KUBE_ROOT}/hack/godep-restore.sh" ${V}

BASH_TARGETS="
	update-generated-protobuf
	update-codegen
	update-generated-runtime
	update-generated-device-plugin
	update-generated-docs
	update-generated-swagger-docs
	update-swagger-spec
	update-openapi-spec
	update-api-reference-docs
	update-staging-godeps
	update-bazel"

for t in ${BASH_TARGETS}; do
	echo -e "${color_yellow}Running $t${color_norm}"
	if ${SILENT} ; then
		if ! bash "${KUBE_ROOT}/hack/$t.sh" 1> /dev/null; then
			echo -e "${color_red}Running $t FAILED${color_norm}"
			if ! ${ALL}; then
				exit 1
			fi
		fi
	else
		if ! bash "${KUBE_ROOT}/hack/$t.sh"; then
			echo -e "${color_red}Running $t FAILED${color_norm}"
			if ! ${ALL}; then
				exit 1
			fi
		fi
	fi
done

echo -e "${color_green}Update scripts completed successfully${color_norm}"
