#!/bin/bash
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
   echo "x86_64"
   wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F"${KUSTOMIZE_VERSION}"/kustomize_"${KUSTOMIZE_VERSION}"_linux_amd64.tar.gz && \
   tar xvf kustomize_"${KUSTOMIZE_VERSION}"_linux_amd64.tar.gz && \
   rm kustomize_"${KUSTOMIZE_VERSION}"_linux_amd64.tar.gz && \
   mv kustomize /usr/bin
elif [ "$ARCH" == "aarch64" ]; then
   echo "arm arch"
   wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F"${KUSTOMIZE_VERSION}"/kustomize_"${KUSTOMIZE_VERSION}"_linux_arm64.tar.gz && \
   tar xvf kustomize_"${KUSTOMIZE_VERSION}"_linux_arm64.tar.gz && \
   rm kustomize_"${KUSTOMIZE_VERSION}"_linux_arm64.tar.gz && \
   mv kustomize /usr/bin
fi
