#!/bin/bash
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
   echo "x86_64"
   wget https://storage.googleapis.com/etcd/"${ETCD_VERSION}"/etcd-"${ETCD_VERSION}"-linux-amd64.tar.gz
   tar xvf etcd-"${ETCD_VERSION}"-linux-amd64.tar.gz && \
   mv etcd-"${ETCD_VERSION}"-linux-amd64/etcd /usr/local/bin/etcd
elif [ "$ARCH" == "aarch64" ]; then
   echo "arm arch"
   wget https://storage.googleapis.com/etcd/"${ETCD_VERSION}"/etcd-"${ETCD_VERSION}"-linux-arm64.tar.gz
   tar xvf etcd-"${ETCD_VERSION}"-linux-arm64.tar.gz && \
   mv etcd-"${ETCD_VERSION}"-linux-arm64/etcd /usr/local/bin/etcd
fi
