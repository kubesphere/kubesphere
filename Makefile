# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.


# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

GV="network:v1alpha1 servicemesh:v1alpha2 tenant:v1alpha1 tenant:v1alpha2 devops:v1alpha1 iam:v1alpha2 devops:v1alpha3 cluster:v1alpha1 storage:v1alpha1 auditing:v1alpha1 types:v1beta1 quota:v1alpha2 application:v1alpha1 notification:v2beta1 gateway:v1alpha1"
MANIFESTS="application/* cluster/* iam/* network/v1alpha1 quota/* storage/* tenant/* gateway/*"

# App Version
APP_VERSION = v3.2.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OUTPUT_DIR=bin
ifeq (${GOFLAGS},)
	# go build with vendor by default.
	export GOFLAGS=-mod=vendor
endif
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
all: test ks-apiserver ks-controller-manager;$(info $(M)...Begin to test and build all of binary.) @ ## Test and build all of binary.

help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: binary
# Build all of binary
binary: | ks-apiserver ks-controller-manager; $(info $(M)...Build all of binary.) @ ## Build all of binary.

# Build ks-apiserver binary
ks-apiserver: ; $(info $(M)...Begin to build ks-apiserver binary.)  @ ## Build ks-apiserver.
	 hack/gobuild.sh cmd/ks-apiserver;

# Build ks-controller-manager binary
ks-controller-manager: ; $(info $(M)...Begin to build ks-controller-manager binary.)  @ ## Build ks-controller-manager.
	hack/gobuild.sh cmd/controller-manager

# Run all verify scripts hack/verify-*.sh
verify-all: ; $(info $(M)...Begin to run all verify scripts hack/verify-*.sh.)  @ ## Run all verify scripts hack/verify-*.sh.
	hack/verify-all.sh

# Build e2e binary
e2e: ;$(info $(M)...Begin to build e2e binary.)  @ ## Build e2e binary.
	hack/build_e2e.sh test/e2e

kind-e2e: ;$(info $(M)...Run e2e test.) @ ## Run e2e test in kind.
	hack/kind_e2e.sh

# Run go fmt against code
fmt: ;$(info $(M)...Begin to run go fmt against code.)  @ ## Run go fmt against code.
	gofmt -w ./pkg ./cmd ./tools ./api

# Format all import, `goimports` is required.
goimports: ;$(info $(M)...Begin to Format all import.)  @ ## Format all import, `goimports` is required.
	@hack/update-goimports.sh

# Run go vet against code
vet: ;$(info $(M)...Begin to run go vet against code.)  @ ## Run go vet against code.
	go vet ./pkg/... ./cmd/...

# Generate manifests e.g. CRD, RBAC etc.
manifests: ;$(info $(M)...Begin to generate manifests e.g. CRD, RBAC etc..)  @ ## Generate manifests e.g. CRD, RBAC etc.
	hack/generate_manifests.sh ${CRD_OPTIONS} ${MANIFESTS}

deploy: manifests ;$(info $(M)...Begin to deploy.)  @ ## Deploy.
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

mockgen: ;$(info $(M)...Begin to mockgen.)  @ ## Mockgen.
	mockgen -package=openpitrix -source=pkg/simple/client/openpitrix/openpitrix.go -destination=pkg/simple/client/openpitrix/mock.go

deepcopy: ;$(info $(M)...Begin to deepcopy.)  @ ## Deepcopy.
	hack/generate_group.sh "deepcopy" kubesphere.io/api kubesphere.io/api ${GV} --output-base=staging/src/  -h "hack/boilerplate.go.txt"

openapi: ;$(info $(M)...Begin to openapi.)  @ ## Openapi.
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/tenant/v1alpha1 -p kubesphere.io/api/tenant/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/network/v1alpha1 -p kubesphere.io/api/network/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/servicemesh/v1alpha2 -p kubesphere.io/api/servicemesh/v1alpha2 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/api/networking/v1,./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/k8s.io/apimachinery/pkg/util/intstr,./vendor/kubesphere.io/api/network/v1alpha1 -p kubesphere.io/api/network/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/devops/v1alpha1,./vendor/k8s.io/apimachinery/pkg/runtime,./vendor/k8s.io/api/core/v1 -p kubesphere.io/api/devops/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/cluster/v1alpha1,./vendor/k8s.io/apimachinery/pkg/runtime,./vendor/k8s.io/api/core/v1 -p kubesphere.io/api/cluster/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/devops/v1alpha3,./vendor/k8s.io/apimachinery/pkg/runtime -p kubesphere.io/api/devops/v1alpha3 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./tools/cmd/crd-doc-gen/main.go
	go run ./tools/cmd/doc-gen/main.go

container: ;$(info $(M)...Begin to build the docker image.)  @ ## Build the docker image.
	DRY_RUN=true hack/docker_build.sh

container-push: ;$(info $(M)...Begin to build and push.)  @ ## Build and Push.
	hack/docker_build.sh

container-cross: ; $(info $(M)...Begin to build container images for multiple platforms.)  @ ## Build container images for multiple platforms. Currently, only linux/amd64,linux/arm64 are supported.
	DRY_RUN=true hack/docker_build_multiarch.sh

container-cross-push: ; $(info $(M)...Begin to build and push.)  @ ## Build and Push.
	hack/docker_build_multiarch.sh

helm-package: ; $(info $(M)...Begin to helm-package.)  @ ## Helm-package.
	ls config/crds/ | xargs -i cp -r config/crds/{} config/ks-core/crds/
	helm package config/ks-core --app-version=${APP_VERSION} --version=0.1.0 -d ./bin

helm-deploy: ; $(info $(M)...Begin to helm-deploy.)  @ ## Helm-deploy.
	ls config/crds/ | xargs -i cp -r config/crds/{} config/ks-core/crds/
	- kubectl create ns kubesphere-controls-system
	helm upgrade --install ks-core ./config/ks-core -n kubesphere-system --create-namespace
	kubectl apply -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/roles/ks-core/prepare/files/ks-init/role-templates.yaml

helm-uninstall: ; $(info $(M)...Begin to helm-uninstall.)  @ ## Helm-uninstall.
	- kubectl delete ns kubesphere-controls-system
	helm uninstall ks-core -n kubesphere-system
	kubectl delete -f https://raw.githubusercontent.com/kubesphere/ks-installer/master/roles/ks-core/prepare/files/ks-init/role-templates.yaml

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: vet test-env ;$(info $(M)...Begin to run tests.)  @ ## Run tests.
	export KUBEBUILDER_ASSETS=$(shell pwd)/testbin/bin; go test ./pkg/... ./cmd/... -covermode=atomic -coverprofile=coverage.txt
	cd staging/src/kubesphere.io/api ; GOFLAGS="" go test ./...
	cd staging/src/kubesphere.io/client-go ; GOFLAGS="" go test ./...

.PHONY: test-env
test-env: ;$(info $(M)...Begin to setup test env) @ ## Download unit test libraries e.g. kube-apiserver etcd.
	@hack/setup-kubebuilder-env.sh

.PHONY: clean
clean: ;$(info $(M)...Begin to clean.)  @ ## Clean.
	-make -C ./pkg/version clean
	@echo "ok"

clientset:  ;$(info $(M)...Begin to find or download controller-gen.)  @ ## Find or download controller-gen,download controller-gen if necessary.
	./hack/generate_client.sh ${GV}

# Fix invalid file's license.
update-licenses: ;$(info $(M)...Begin to update licenses.)
	@hack/update-licenses.sh
