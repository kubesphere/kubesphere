#!/usr/bin/env bash

# This is a modified version of Kubernetes
KUBE_GO_PACKAGE=kubesphere.io/kubesphere

# Ensure the go tool exists and is a viable version.
kube::golang::verify_go_version() {
  if [[ -z "$(command -v go)" ]]; then
    kube::log::usage_from_stdin <<EOF
Can't find 'go' in PATH, please fix and retry.
See http://golang.org/doc/install for installation instructions.
EOF
    return 2
  fi

  local go_version
  IFS=" " read -ra go_version <<< "$(go version)"
  local minimum_go_version
  minimum_go_version=go1.20
  if [[ "${minimum_go_version}" != $(echo -e "${minimum_go_version}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && "${go_version[2]}" != "devel" ]]; then
    kube::log::usage_from_stdin <<EOF
Detected go version: ${go_version[*]}.
Kubernetes requires ${minimum_go_version} or greater.
Please install ${minimum_go_version} or later.
EOF
    return 2
  fi
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
# in order to set the Kubernetes based on the git tree status.
# IMPORTANT: if you update any of these, also update the lists in
# pkg/version/def.bzl and hack/print-workspace-status.sh.
kube::version::ldflags() {
  kube::version::get_version_vars

  local -a ldflags
  function add_ldflag() {
    local key=${1}
    local val=${2}
    # If you update these, also update the list component-base/version/def.bzl.
    ldflags+=(
      "-X '${KUBE_GO_PACKAGE}/pkg/version.${key}=${val}'"
    )
  }

  add_ldflag "buildDate" "$(date ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} -u +'%Y-%m-%dT%H:%M:%SZ')"
  if [[ -n ${KUBE_GIT_COMMIT-} ]]; then
    add_ldflag "gitCommit" "${KUBE_GIT_COMMIT}"
    add_ldflag "gitTreeState" "${KUBE_GIT_TREE_STATE}"
  fi

  if [[ -n ${KUBE_GIT_VERSION-} ]]; then
    add_ldflag "gitVersion" "${KUBE_GIT_VERSION}"
  fi

  if [[ -n ${KUBE_GIT_MAJOR-} && -n ${KUBE_GIT_MINOR-} ]]; then
    add_ldflag "gitMajor" "${KUBE_GIT_MAJOR}"
    add_ldflag "gitMinor" "${KUBE_GIT_MINOR}"
  fi

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}
