#!/usr/bin/env bash

CRDS_PATH=$1
echo "ks-crds pre upgrade..."
for crd in "$CRDS_PATH"/*.yaml; do
  basename "$crd"
  kubectl apply -f "$crd"
done