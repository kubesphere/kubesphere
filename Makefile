# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# The binary to build 
BIN ?= ks-apiserver

IMG ?= kubespheredev/ks-apiserver
OUTPUT_DIR=bin

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
all: test ks-apiserver ks-apigateway ks-iam

# Build ks-apiserver binary
ks-apiserver: test
	hack/gobuild.sh cmd/ks-apiserver

# Build ks-apigateway binary
ks-apigateway: test
	hack/gobuild.sh cmd/ks-apigateway

# Build ks-iam binary
ks-iam: test
	hack/gobuild.sh cmd/ks-iam

# Run go fmt against code 
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate DeepCopy to implement runtime.Object
deepcopy:
	./vendor/k8s.io/code-generator/generate-groups.sh deepcopy kubesphere.io/kubesphere/pkg/client kubesphere.io/kubesphere/pkg/apis "servicemesh:v1alpha2"

# Generate code
generate:
ifndef GOPATH
	$(error GOPATH not defined, please define GOPATH. Run "go help gopath" to learn more about GOPATH)
endif
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: all
	docker build . -t ${IMG}

# Run tests
test: generate fmt vet
	export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=1m; go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: clean
clean:
	-make -C ./pkg/version clean
	@echo "ok"
