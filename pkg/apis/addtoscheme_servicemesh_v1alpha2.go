/*
Copyright 2019 The KubeSphere Authors.

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

package apis

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"

	appv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha2.SchemeBuilder.AddToScheme)

	// Register networking.istio.io/v1alpha3
	AddToSchemes = append(AddToSchemes, v1alpha3.SchemeBuilder.AddToScheme)

	// Register application scheme
	AddToSchemes = append(AddToSchemes, appv1beta1.SchemeBuilder.AddToScheme)
}
