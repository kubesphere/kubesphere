# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

GV="network:v1alpha1 servicemesh:v1alpha2 tenant:v1alpha1 tenant:v1alpha2 devops:v1alpha1 iam:v1alpha2 devops:v1alpha3 cluster:v1alpha1 storage:v1alpha1 auditing:v1alpha1 types:v1beta1 quota:v1alpha2 application:v1alpha1 notification:v2beta1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OUTPUT_DIR=bin
GOFLAGS=-mod=vendor
define ALL_HELP_INFO
# Build code.
#
# Args:
#   WHAT: Directory names to build.  If any of these directories has a 'main'
#     package, the build will produce executable files under $(OUT_DIR).
#     If not specified, "everything" will be built.
#   GOFLAGS: Extra flags to pass to 'go' when building.
#   GOLDFLAGS: Extra linking flags passed to 'go' when building.
#   GOGCFLAGS: Additional go compile flags passed to 'go' when building.
#
# Example:
#   make
#   make all
#   make all WHAT=cmd/ks-apiserver
#     Note: Use the -N -l options to disable compiler optimizations an inlining.
#           Using these build options allows you to subsequently use source
#           debugging tools like delve.
endef
.PHONY: all
all: test ks-apiserver ks-controller-manager

# Build ks-apiserver binary
ks-apiserver:
	hack/gobuild.sh cmd/ks-apiserver

# Build ks-controller-manager binary
ks-controller-manager:
	hack/gobuild.sh cmd/controller-manager

# Run all verify scripts hack/verify-*.sh
verify-all:
	hack/verify-all.sh

# Build e2e binary
e2e:
	hack/build_e2e.sh test/e2e

# Run go fmt against code 
fmt:
	gofmt -w ./pkg ./cmd ./tools ./api

# Format all import, `goimports` is required.
goimports:
	@hack/update-goimports.sh

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/application/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/cluster/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/devops/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/iam/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/network/v1alpha1/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/quota/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/storage/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds
	go run ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=./hack/boilerplate.go.txt paths=kubesphere.io/api/tenant/... rbac:roleName=controller-perms ${CRD_OPTIONS} output:crd:artifacts:config=config/crds

deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

mockgen:
	mockgen -package=openpitrix -source=pkg/simple/client/openpitrix/openpitrix.go -destination=pkg/simple/client/openpitrix/mock.go

deepcopy:
	hack/generate_group.sh "deepcopy" kubesphere.io/api kubesphere.io/api ${GV} --output-base=staging/src/  -h "hack/boilerplate.go.txt"

openapi:
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./staging/src/kubesphere.io/api/tenant/v1alpha1 -p kubesphere.io/api/tenant/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./staging/src/kubesphere.io/api/servicemesh/v1alpha2 -p kubesphere.io/api/servicemesh/v1alpha2 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/api/networking/v1,./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/k8s.io/apimachinery/pkg/util/intstr,./staging/src/kubesphere.io/api/network/v1alpha1 -p kubesphere.io/api/network/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./staging/src/kubesphere.io/api/devops/v1alpha1,./vendor/k8s.io/apimachinery/pkg/runtime,./vendor/k8s.io/api/core/v1 -p kubesphere.io/api/devops/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./staging/src/kubesphere.io/api/cluster/v1alpha1,./vendor/k8s.io/apimachinery/pkg/runtime,./vendor/k8s.io/api/core/v1 -p kubesphere.io/api/cluster/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./staging/src/kubesphere.io/api/devops/v1alpha3,./vendor/k8s.io/apimachinery/pkg/runtime -p kubesphere.io/api/devops/v1alpha3 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list
	go run ./tools/cmd/crd-doc-gen/main.go
	go run ./tools/cmd/doc-gen/main.go

# Build the docker image
container:
	DRY_RUN=true hack/docker_build.sh

# Build and push 
container-push:
	hack/docker_build.sh

# Build container images for multiple platforms.
#   Currently, only linux/amd64,linux/arm64 are supported.
container-cross:
	DRY_RUN=true hack/docker_build_multiarch.sh

# Build and push
container-cross-push:
	hack/docker_build_multiarch.sh

helm-package:
	ls config/crds/ | xargs -i cp -r config/crds/{} config/ks-core/crds/
	helm package config/ks-core --app-version=v3.1.0 --version=0.1.0 -d ./bin

helm-deploy:
	ls config/crds/ | xargs -i cp -r config/crds/{} config/ks-core/crds/
	- kubectl create ns kubesphere-controls-system
	helm upgrade --install ks-core ./config/ks-core -n kubesphere-system --create-namespace
	kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/roles/ks-core/prepare/files/ks-init/role-templates.yaml

helm-uninstall:
	- kubectl delete ns kubesphere-controls-system
	helm uninstall ks-core -n kubesphere-system
	kubectl delete -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/roles/ks-core/prepare/files/ks-init/role-templates.yaml

# Run tests
test: vet
	export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m; go test ./pkg/... ./cmd/... -covermode=atomic -coverprofile=coverage.txt

.PHONY: clean
clean:
	-make -C ./pkg/version clean
	@echo "ok"

# find or download controller-gen
# download controller-gen if necessary
clientset: 
	./hack/generate_client.sh ${GV}
