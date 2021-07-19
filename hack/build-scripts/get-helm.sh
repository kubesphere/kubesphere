#!/bin/bash
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
   echo "x86_64"
   wget https://get.helm.sh/helm-"${HELM_VERSION}"-linux-amd64.tar.gz && \
   tar xvf helm-"${HELM_VERSION}"-linux-amd64.tar.gz && \
   rm helm-"${HELM_VERSION}"-linux-amd64.tar.gz && \
   mv linux-amd64/helm /usr/bin/ && \
   rm -rf linux-amd64
elif [ "$ARCH" == "aarch64" ]; then
   echo "arm arch"
   wget https://get.helm.sh/helm-"${HELM_VERSION}"-linux-arm64.tar.gz && \
   tar xvf helm-"${HELM_VERSION}"-linux-arm64.tar.gz && \
   rm helm-"${HELM_VERSION}"-linux-arm64.tar.gz && \
   mv linux-arm64/helm /usr/bin/ && \
   rm -rf linux-arm64
fi
