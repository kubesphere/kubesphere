#!/bin/bash

set -e

GV="network:v1alpha1 servicemesh:v1alpha2 tenant:v1alpha1 tenant:v1alpha2 devops:v1alpha1 iam:v1alpha2 devops:v1alpha3 cluster:v1alpha1 storage:v1alpha1 auditing:v1alpha1 types:v1beta1"

rm -rf ./pkg/client
./hack/generate_group.sh "client,lister,informer" kubesphere.io/kubesphere/pkg/client kubesphere.io/kubesphere/pkg/apis "$GV" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
mv kubesphere.io/kubesphere/pkg/client ./pkg/
rm -rf ./kubesphere.io
