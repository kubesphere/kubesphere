# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:allowDangerousTypes=true"
MANIFESTS="cluster/v1alpha1 iam/... quota/v1alpha2 storage/v1alpha1 tenant/... extensions/v1alpha1 core/v1alpha1 gateway/v1alpha2 application/v2"

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
	gofmt -w ./pkg ./cmd ./tools ./api ./staging ./kube

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

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
deepcopy: ;$(info $(M)...Begin to deepcopy.)  @ ## Deepcopy.
	hack/generate_manifests.sh ${CRD_OPTIONS} ${MANIFESTS} "deepcopy"

openapi: ;$(info $(M)...Begin to openapi.)  @ ## Openapi.
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/tenant/v1beta1 -p kubesphere.io/api/tenant/v1beta1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
	go run ./vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./vendor/kubesphere.io/api/cluster/v1alpha1,./vendor/k8s.io/apimachinery/pkg/runtime,./vendor/k8s.io/api/core/v1 -p kubesphere.io/api/cluster/v1alpha1 -h ./hack/boilerplate.go.txt --report-filename ./api/api-rules/violation_exceptions.list  --output-base=staging/src/
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
	helm package config/ks-core -d ./bin

helm-deploy: ; $(info $(M)...Begin to helm-deploy.)  @ ## Helm-deploy.
	- kubectl create ns kubesphere-controls-system
	helm upgrade --install ks-core ./config/ks-core -n kubesphere-system --create-namespace

helm-uninstall: ; $(info $(M)...Begin to helm-uninstall.)  @ ## Helm-uninstall.
	- kubectl delete ns kubesphere-controls-system
	helm uninstall ks-core -n kubesphere-system

# Run tests
test: vet test-env ;$(info $(M)...Begin to run tests.)  @ ## Run tests.
	go test ./pkg/... ./cmd/... -covermode=atomic -coverprofile=coverage.txt
	cd staging/src/kubesphere.io/api ; GOFLAGS="" go test ./...
	cd staging/src/kubesphere.io/client-go ; GOFLAGS="" go test ./...

.PHONY: test-env
test-env: ;$(info $(M)...Begin to setup test env) @ ## Download unit test libraries e.g. kube-apiserver etcd.
	@hack/setup-envtest.sh

.PHONY: clean
clean: ;$(info $(M)...Begin to clean.)  @ ## Clean.
	-make -C ./pkg/version clean
	@echo "ok"

# Fix invalid file's license.
update-licenses: ;$(info $(M)...Begin to update licenses.)
	@hack/update-licenses.sh

golint:
	@hack/verify-golangci-lint.sh
