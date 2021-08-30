#!/bin/bash
# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL="/usr/bin/env bash -o pipefail"

ENVTEST_ASSETS_DIR=$(pwd)/testbin
mkdir -p "${ENVTEST_ASSETS_DIR}"
test -f "${ENVTEST_ASSETS_DIR}"/setup-envtest.sh || curl -sSLo "${ENVTEST_ASSETS_DIR}"/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
source "${ENVTEST_ASSETS_DIR}"/setup-envtest.sh; fetch_envtest_tools "${ENVTEST_ASSETS_DIR}"; setup_envtest_env "${ENVTEST_ASSETS_DIR}"
