// This is a generated file. Do not edit directly.
// Run hack/pin-dependency.sh to change pinned dependency versions.
// Run hack/update-vendor.sh to update go.mod files and the vendor directory.

module kubesphere.io/api

go 1.13

require (
	github.com/go-openapi/spec v0.19.3
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3
	github.com/projectcalico/libcalico-go v1.7.2-0.20191014160346-2382c6cdd056
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	istio.io/api v0.0.0-20200812202721-24be265d41c3
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/application v0.8.3
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.4
)
