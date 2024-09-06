#!/usr/bin/env bash

CRDS_PATH=$1
echo "ks-crds pre upgrade..."
# shellcheck disable=SC1060
for crd in `ls $CRDS_PATH|grep \.yaml$`; do
  echo $crd
  kubectl apply -f $CRDS_PATH/$crd
done