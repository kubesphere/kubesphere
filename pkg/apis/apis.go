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

// Generate deepcopy for apis
//go:generate go run ../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../hack/boilerplate.go.txt

// Package apis contains Kubernetes API groups.
package apis

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/apibuilder"
	metrics "kubesphere.io/kubesphere/pkg/apis/metrics/install"
	operations "kubesphere.io/kubesphere/pkg/apis/operations/install"
	resources "kubesphere.io/kubesphere/pkg/apis/resources/install"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}

var APIBuilder apibuilder.APIBuilder

func init() {
	APIBuilder = append(APIBuilder, metrics.Install, resources.Install, operations.Install)
}

func AddToContainer(container *restful.Container) {
	APIBuilder.AddToContainer(container)
}
