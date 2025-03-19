#!/usr/bin/env bash

CRD_NAMES=$1
MAPPING_CONFIG=$2

for extension in $(kubectl get installplan -o json | jq -r '.items[] | select(.status.state == "Installed") | .metadata.name'); do
  namespace=$(kubectl get installplan "$extension" -o=jsonpath='{.status.targetNamespace}')
  version=$(kubectl get extension "$extension" -o=jsonpath='{.status.installedVersion}')
  extensionVersion="$extension-$version"
  echo "Found extension $extensionVersion installed"
  if helm status "$extension" --namespace "$namespace" &>/dev/null; then
    helm mapkubeapis "$extension" --namespace "$namespace" --mapfile "$MAPPING_CONFIG"
  fi
  if helm status "$extension-agent" --namespace "$namespace" &>/dev/null; then
    helm mapkubeapis "$extension-agent" --namespace "$namespace" --mapfile "$MAPPING_CONFIG"
  fi
done

kubectl patch workspaces.tenant.kubesphere.io system-workspace -p '{"metadata":{"finalizers":[]}}' --type=merge
kubectl patch workspacetemplates.tenant.kubesphere.io system-workspace -p '{"metadata":{"finalizers":[]}}' --type=merge

for ns in $(kubectl get ns -o jsonpath='{.items[*].metadata.name}' -l 'kubesphere.io/managed=true'); do
  kubectl label ns "$ns" kubesphere.io/workspace- && \
  kubectl patch ns "$ns" -p '{"metadata":{"ownerReferences":[]}}' --type=merge && \
  echo '{"kind":"Namespace","apiVersion":"v1","metadata":{"name":"'"$ns"'","finalizers":null}}' | kubectl replace --raw "/api/v1/namespaces/$ns/finalize" -f -
done

for crd in $(kubectl get crds -o jsonpath='{.items[*].metadata.name}'); do
  if [[ " ${CRD_NAMES[*]} " =~ ${crd} ]]; then
    echo "Deleting CRD $crd"
    scope=$(kubectl get crd "$crd" -o jsonpath='{.spec.scope}')
    if [[ $scope == "Namespaced" ]]; then
      kubectl get "$crd" -A --no-headers | awk '{print $1" "$2" ""'"$crd"'"}' | xargs -n 3 sh -c 'kubectl patch $2 -n $0 $1 -p "{\"metadata\":{\"finalizers\":null}}" --type=merge 2>/dev/null && kubectl delete $2 -n $0 $1 2>/dev/null'
    else
      kubectl get "$crd" -A --no-headers | awk '{print $1" ""'"$crd"'"}' | xargs -n 2 sh -c 'kubectl patch $1 $0 -p "{\"metadata\":{\"finalizers\":null}}" --type=merge 2>/dev/null && kubectl delete $1 $0 2>/dev/null'
    fi
    kubectl delete crd "$crd" 2>/dev/null
  fi
done