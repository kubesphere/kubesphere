#!/bin/bash
ARCH=$(uname -m)
if [ "$ARCH" == "aarch64" ]; then
  export ETCD_UNSUPPORTED_ARCH=arm64
fi
