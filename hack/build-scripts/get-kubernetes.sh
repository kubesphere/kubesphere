#!/bin/bash
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
   echo "x86_64"
   wget https://dl.k8s.io/"${KUBE_VERSION}"/kubernetes-server-linux-amd64.tar.gz && \
   tar xvf kubernetes-server-linux-amd64.tar.gz &&
   mv kubernetes /usr/local/
elif [ "$ARCH" == "aarch64" ]; then
   echo "arm arch"
   wget https://dl.k8s.io/"${KUBE_VERSION}"/kubernetes-server-linux-arm64.tar.gz && \
   tar xvf kubernetes-server-linux-arm64.tar.gz &&
   mv kubernetes /usr/local/
fi
