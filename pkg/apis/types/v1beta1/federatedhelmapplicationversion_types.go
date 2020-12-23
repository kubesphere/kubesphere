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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"strings"
)

const (
	ResourcePluralFederatedHelmApplicationVersion = "federatedhelmapplicationversions"
	FederatedHelmApplicationVersionKind           = "FederatedHelmApplicationVersion"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FederatedHelmApplicationVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedHelmApplicationVersionSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedHelmApplicationVersionSpec struct {
	Template  HelmApplicationVersionTemplate `json:"template"`
	Placement GenericPlacementFields         `json:"placement"`
	Overrides []GenericOverrideItem          `json:"overrides,omitempty"`
}

type HelmApplicationVersionTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v1alpha1.HelmApplicationVersionSpec `json:"spec,omitempty"`
	AuditSpec         v1alpha1.HelmAuditSpec              `json:"auditSpec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FederatedHelmApplicationVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedHelmApplicationVersion `json:"items"`
}

func (in *FederatedHelmApplicationVersion) GetCreator() string {
	if in == nil || in.Annotations == nil {
		return ""
	}
	return in.Annotations[constants.CreatorAnnotationKey]
}

func (in *FederatedHelmApplicationVersion) GetHelmApplicationVersionId() string {
	if in == nil {
		return ""
	}
	return in.Name
}

func (in *FederatedHelmApplicationVersion) GetWorkspace() string {
	if in == nil {
		return ""
	}
	if l := in.Labels; l == nil {
		return ""
	} else {
		return l[constants.WorkspaceLabelKey]
	}
}

func (in *FederatedHelmApplicationVersion) GetVersionName() string {
	if in == nil {
		return ""
	}

	appV := in.GetChartAppVersion()
	if appV != "" {
		return fmt.Sprintf("%s [%s]", in.GetChartVersion(), appV)
	} else {
		return in.GetChartVersion()
	}
}

func (in *FederatedHelmApplicationVersion) GetHelmApplicationId() string {
	if in == nil {
		return ""
	}
	if l := in.Labels; l == nil {
		return ""
	} else {
		return l[constants.ChartApplicationIdLabelKey]
	}
}

func (in *FederatedHelmApplicationVersion) GetSemver() string {
	return strings.Split(in.GetVersionName(), " ")[0]
}

func (in *FederatedHelmApplicationVersion) GetTrueName() string {
	if in == nil {
		return ""
	}
	return in.Spec.Template.Spec.Name
}

func (in *FederatedHelmApplicationVersion) GetChartVersion() string {
	if in == nil {
		return ""
	}

	return in.Spec.Template.Spec.Version
}

func (in *FederatedHelmApplicationVersion) GetChartAppVersion() string {
	if in == nil {
		return ""
	}

	return in.Spec.Template.Spec.AppVersion
}

func (in *FederatedHelmApplicationVersion) GetHelmRepoId() string {
	if in == nil {
		return ""
	}

	if l := in.Labels; l == nil {
		return ""
	} else {
		return l[constants.ChartRepoIdLabelKey]
	}
}
