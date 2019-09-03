/*
Copyright 2019 The KubeSphere authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package apis contains KubeSphere API groups.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Generate openapi for apis
//go:generate go run ../../vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ../../vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./tenant/v1alpha1 -p kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1 -h ../../hack/boilerplate.go.txt --report-filename ../../api/api-rules/violation_exceptions.list
//go:generate go run ../../vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ../../vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./servicemesh/v1alpha2 -p kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2 -h ../../hack/boilerplate.go.txt --report-filename ../../api/api-rules/violation_exceptions.list
//go:generate go run ../../vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ../../vendor/k8s.io/api/networking/v1,../../vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./network/v1alpha1 -p kubesphere.io/kubesphere/pkg/apis/network/v1alpha1 -h ../../hack/boilerplate.go.txt --report-filename ../../api/api-rules/violation_exceptions.list
//go:generate go run ../../vendor/k8s.io/kube-openapi/cmd/openapi-gen/openapi-gen.go -O openapi_generated -i ../../vendor/k8s.io/apimachinery/pkg/apis/meta/v1,./devops/v1alpha1 -p kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1 -h ../../hack/boilerplate.go.txt --report-filename ../../api/api-rules/violation_exceptions.list

// Generate deepcopy for apis

//go:generate ${GOPATH}/bin/deepcopy-gen -i kubesphere.io/kubesphere/pkg/apis/... -h ../../hack/boilerplate.go.txt -O zz_generated.deepcopy
// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
