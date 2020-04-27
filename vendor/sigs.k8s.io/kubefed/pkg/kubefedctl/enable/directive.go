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

package enable

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
)

// EnableTypeDirectiveSpec defines the desired state of EnableTypeDirective.
type EnableTypeDirectiveSpec struct {
	// The API version of the target type.
	// +optional
	TargetVersion string `json:"targetVersion,omitempty"`

	// The name of the API group to use for generated federated types.
	// +optional
	FederatedGroup string `json:"federatedGroup,omitempty"`

	// The API version to use for generated federated types.
	// +optional
	FederatedVersion string `json:"federatedVersion,omitempty"`
}

// TODO(marun) This should become a proper API type and drive enabling
// type federation via a controller.  For now its only purpose is to
// enable loading of configuration from disk.
type EnableTypeDirective struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EnableTypeDirectiveSpec `json:"spec,omitempty"`
}

func (ft *EnableTypeDirective) SetDefaults() {
	ft.Spec.FederatedGroup = options.DefaultFederatedGroup
	ft.Spec.FederatedVersion = options.DefaultFederatedVersion
}

func NewEnableTypeDirective() *EnableTypeDirective {
	ft := &EnableTypeDirective{}
	ft.SetDefaults()
	return ft
}
