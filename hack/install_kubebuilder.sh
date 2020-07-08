#!/bin/sh

#
# This file will be  fetched as: curl -L https://git.io/getLatestKubebuilder | sh -
# so it should be pure bourne shell, not bash (and not reference other scripts)
#
# The script fetches the latest kubebuilder release candidate and untars it.
# It lets users to do curl -L https://git.io//getLatestKubebuilder | KUBEBUILDER_VERSION=1.0.5 sh -
# for instance to change the version fetched.

# Check if the program is installed, otherwise exit
function command_exists () {
  if ! [[ -x "$(command -v $1)" ]]; then
    echo "Error: $1 program is not installed." >&2
    exit 1
  fi
}

# Determine OS
OS="$(uname)"
case $OS in
  Darwin)
    OSEXT="darwin"
    ;;
  Linux)
    OSEXT="linux"
    ;;
  *)
    echo "Only OSX and Linux OS are supported !"
    exit 1
    ;;
esac

HW=$(uname -m)
case $HW in
    x86_64)
      ARCH=amd64 ;;
    *)
      echo "Only x86_64 machines are supported !"
      exit 1
      ;;
esac

# Check if curl, tar commands/programs exist
command_exists curl
command_exists tar

KUBEBUILDER_VERSION=v2.3.1
KUBEBUILDER_VERSION=${KUBEBUILDER_VERSION#"v"}
KUBEBUILDER_VERSION_NAME="kubebuilder_${KUBEBUILDER_VERSION}"
KUBEBUILDER_DIR=/usr/local/kubebuilder

# Check if folder containing kubebuilder executable exists and is not empty
if [[ -d "$KUBEBUILDER_DIR" ]]; then
  if [[ "$(ls -A ${KUBEBUILDER_DIR})" ]]; then
    echo "\n/usr/local/kubebuilder folder is not empty. Please delete or backup it before to install ${KUBEBUILDER_VERSION_NAME}"
    exit 1
  fi
fi

TMP_DIR=$(mktemp -d)
pushd $TMP_DIR

# Downloading Kubebuilder compressed file using curl program
URL="https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/${KUBEBUILDER_VERSION_NAME}_${OSEXT}_${ARCH}.tar.gz"
echo "Downloading ${KUBEBUILDER_VERSION_NAME}\nfrom $URL\n"
curl -L "$URL"| tar xz -C ${TMP_DIR}

echo "Downloaded executable files"
ls "${KUBEBUILDER_VERSION_NAME}_${OSEXT}_${ARCH}/bin"

echo "Moving files to $KUBEBUILDER_DIR folder\n"
mv ${KUBEBUILDER_VERSION_NAME}_${OSEXT}_${ARCH} kubebuilder && sudo mv -f kubebuilder /usr/local/

echo "Add kubebuilder to your path; e.g copy paste in your shell and/or edit your ~/.profile file"
echo "export PATH=\$PATH:/usr/local/kubebuilder/bin"
popd
rm -rf ${TMP_DIR}

export PATH=$PATH:/usr/local/kubebuilder/bin
