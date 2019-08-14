#!/bin/bash

set -e

function cleanup(){
    result=$?
    set +e
    echo "Cleaning Namespace"
    kubectl delete ns $TEST_NS > /dev/null
    exit $result
}

tag=`git rev-parse --short HEAD`
IMG=magicsong/ks-network:$tag
DEST=/tmp/manager.yaml
TEST_NS=porter-test-$tag
SKIP_BUILD=no

##cleanning before running
kubectl get ns $TEST_NS 2>&1 | grep "not found" || kubectl delete ns $TEST_NS

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
    --default)
    DEFAULT=YES
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done

trap cleanup EXIT SIGINT SIGQUIT
kustomize_dir="./kustomize/network"
if [ "$(uname)" == "Darwin" ]; then
    sed -i '' -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' $kustomize_dir/kustomization.yaml
    sed -i '' -e 's@image: .*@image: '"${IMG}"'@' $kustomize_dir/patch_image_name.yaml
else
    sed -i  -e  's/namespace: .*/namespace: '"${TEST_NS}"'/' $kustomize_dir/kustomization.yaml
    sed -i -e 's@image: .*@image: '"${IMG}"'@' $kustomize_dir/patch_image_name.yaml
fi

kubectl create ns  $TEST_NS
kubectl apply -k $kustomize_dir
