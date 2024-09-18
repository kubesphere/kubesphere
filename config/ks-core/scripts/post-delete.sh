#!/usr/bin/env bash

# set -x

EXTENSION_RELATED_RESOURCES='jobs.batch roles.rbac.authorization.k8s.io rolebindings.rbac.authorization.k8s.io clusterroles.rbac.authorization.k8s.io clusterrolebindings.rbac.authorization.k8s.io'

for resource in $EXTENSION_RELATED_RESOURCES;do
  echo "kubectl delete $resource -l kubesphere.io/extension-ref --all-namespaces"
  kubectl delete $resource -l kubesphere.io/managed=true --all-namespaces
done
