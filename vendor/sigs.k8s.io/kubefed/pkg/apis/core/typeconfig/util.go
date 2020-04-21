/*
Copyright 2018 The Kubernetes Authors.

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

package typeconfig

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupQualifiedName returns the plural name of the api resource
// optionally qualified by its group:
//
//   '<target plural name>[.<target group name>]'
//
// This is the naming scheme for FederatedTypeConfig resources.  The
// scheme ensures that, for a given KubeFed control plane,
// federation of a target type will be configured by at most one
// FederatedTypeConfig.
func GroupQualifiedName(apiResource metav1.APIResource) string {
	if len(apiResource.Group) == 0 {
		return apiResource.Name
	}
	return fmt.Sprintf("%s.%s", apiResource.Name, apiResource.Group)
}
