#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function wait_for_installation_finish() {
    echo "waiting for ks-installer pod ready"
    kubectl -n kubesphere-system wait --timeout=180s --for=condition=Ready "$(kubectl -n kubesphere-system get pod -l app=ks-install -oname)"
    echo "waiting for KubeSphere ready"
    while IFS= read -r line; do
        if [[ $line =~ "Welcome to KubeSphere" ]]
            then
                break
        fi
    done < <(timeout 900 kubectl logs -n kubesphere-system deploy/ks-installer -f)
}

# Use kubespheredev and latest tag as default image
TAG="${TAG:-latest}"
REPO="${REPO:-kubespheredev}"

# Use KIND_LOAD_IMAGE=y .hack/deploy-kubesphere.sh to load
# the built docker image into kind before deploying.
if [[ "${KIND_LOAD_IMAGE:-}" == "y" ]]; then
    kind load docker-image "$REPO/ks-apiserver:$TAG" --name="${KIND_CLUSTER_NAME:-kind}"
    kind load docker-image "$REPO/ks-controller-manager:$TAG" --name="${KIND_CLUSTER_NAME:-kind}"
fi

#TODO: override ks-apiserver and ks-controller-manager images with specific tag
kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/deploy/kubesphere-installer.yaml
kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/deploy/cluster-configuration.yaml


wait_for_installation_finish

# Expose KubeSphere API Server
kubectl -n kubesphere-system patch svc ks-apiserver -p '{"spec":{"type":"NodePort","ports":[{"name":"ks-apiserver","port":80,"protocol":"TCP","targetPort":9090,"nodePort":30881}]}}'
