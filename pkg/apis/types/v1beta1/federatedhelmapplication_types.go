/*
Copyright 2020 KubeSphere Authors

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"strings"
)

const (
	ResourcePluralFederatedHelmApplication = "federatedhelmapplications"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FederatedHelmApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedHelmApplicationSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedHelmApplicationSpec struct {
	Template  HelmApplicationTemplate `json:"template"`
	Placement GenericPlacementFields  `json:"placement"`
	Overrides []GenericOverrideItem   `json:"overrides,omitempty"`
}

type HelmApplicationTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v1alpha1.HelmApplicationSpec `json:"spec,omitempty"`
	//Don't save into etcd, just used in memory
	AuditSpec v1alpha1.HelmAuditSpec `json:"auditSpec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FederatedHelmApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedHelmApplication `json:"items"`
}

func (in *FederatedHelmApplication) GetTrueName() string {
	if in == nil {
		return ""
	}
	return in.Spec.Template.Spec.Name
}

func (in *FederatedHelmApplication) GetHelmRepoId() string {
	if in == nil {
		return ""
	}

	return getValue(in.Labels, constants.ChartRepoIdLabelKey)
}

func (in *FederatedHelmApplication) GetCreator() string {
	if in == nil || in.Annotations == nil {
		return ""
	}
	return in.Annotations[constants.CreatorAnnotationKey]
}

func (in *FederatedHelmApplication) GetHelmApplicationId() string {
	if in == nil {
		return ""
	}
	return strings.TrimSuffix(in.Name, constants.HelmApplicationAppStoreSuffix)
}
func (in *FederatedHelmApplication) GetHelmCategoryId() string {
	if in == nil {
		return ""
	}

	return getValue(in.Labels, constants.CategoryIdLabelKey)
}

func (in *FederatedHelmApplication) GetWorkspace() string {
	if in == nil {
		return ""
	}

	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func getValue(m map[string]string, key string) string {
	if m == nil {
		return ""
	} else {
		return m[key]
	}
}

func (in *FederatedHelmApplication) GetCategoryId() string {
	if in == nil {
		return ""
	}
	return getValue(in.Labels, constants.CategoryIdLabelKey)
}
