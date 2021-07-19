#!/bin/bash

set -e

workspace=$(pwd)
tag=$(git rev-parse --short HEAD)
IMG=kubespheredev/ks-network:$tag
DEST=/tmp/manager.yaml
TEST_NS=network-test-$tag
SKIP_BUILD=no
STORE_MODE=etcd
MODE="test"

export TEST_NAMESPACE=$TEST_NS
export YAML_PATH=$DEST
export CRD_PATH=$workspace/kustomize/crds
export DEPLOY_NAME=network-manager

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -s|--skip-build)
    SKIP_BUILD=yes
    shift # past argument
    ;;
    -n|--NAMESPACE)
    TEST_NS=$2
    shift # past argument
    shift # past value
    ;;
    -t|--tag)
    tag="$2"
    shift # past argument
    shift # past value
    ;;
    -S|--store-mode)
    STORE_MODE="$2"
    shift # past argument
    shift # past value
    ;;
    -m|--mode)
    MODE="$2"
    shift # past argument
    shift # past value
    ;;
    --default)
    # DEFAULT=YES
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
if [ $SKIP_BUILD == "no" ]; then
    echo "Building binary"
    hack/gobuild.sh cmd/ks-network
    docker build -f build/ks-network/Dockerfile -t "$IMG" bin/cmd
    echo "Push images"
    docker push "$IMG"
fi

kustomize_dir="./kustomize/network/calico-${STORE_MODE}"
if [ "$(uname)" == "Darwin" ]; then
    sed -i '' -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' "$kustomize_dir"/kustomization.yaml
    sed -i '' -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' "$kustomize_dir"/patch_role_binding.yaml
    sed -i '' -e 's@image: .*@image: '"${IMG}"'@' "$kustomize_dir"/patch_image_name.yaml
else
    sed -i  -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' "$kustomize_dir"/patch_role_binding.yaml
    sed -i  -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' "$kustomize_dir"/kustomization.yaml
    sed -i -e 's@image: .*@image: '"${IMG}"'@' "$kustomize_dir"/patch_image_name.yaml
fi

kustomize build "$kustomize_dir" -o $DEST
if [ "$MODE" == "test" ]; then
    ginkgo -v ./test/e2e/...
elif  [ "$MODE" == "debug" ]; then
    kubectl create ns "$TEST_NS" --dry-run -o yaml | kubectl apply -f -
    kubectl apply -f $DEST
fi

