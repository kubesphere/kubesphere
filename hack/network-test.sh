#!/bin/bash

#this script must invoked in the root directory of this repo


tag=`git rev-parse --short HEAD`
IMG=magicsong/ks-network:$tag
DEST=/tmp/manager.yaml
SKIP_BUILD=no

echo "try to delete old yaml"
kubectl delete -f $DEST
set -e
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

if [ $SKIP_BUILD == "no" ]; then
    echo "Building binary"
    hack/gobuild.sh cmd/ks-network
    docker build -f build/ks-network/Dockerfile -t $IMG bin/cmd
    echo "Push images"
    docker push $IMG
fi

echo "Generating yaml"
sed -e 's@image: .*@image: '"${IMG}"'@' config/manager/network.yaml > $DEST
kubectl apply -f $DEST
kubectl apply -f config/rbac/rbac_role_binding_network.yaml


