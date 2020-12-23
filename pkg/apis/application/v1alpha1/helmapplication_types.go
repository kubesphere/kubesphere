/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/constants"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelmApplicationSpec defines the desired state of HelmApplication
type HelmApplicationSpec struct {
	//chart name
	Name string `json:"name"`
	//chart description
	Description string `json:"description,omitempty"`
	//attachments id
	Attachments []string `json:"attachments,omitempty"`
	Abstraction string   `json:"abstraction,omitempty"`
	//application status
	Status  string `json:"status"`
	AppHome string `json:"appHome,omitempty"`
	// The attachment id of the icon
	Icon       string       `json:"icon,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	//the time when status changed
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

// HelmApplicationStatus defines the observed state of HelmApplication

type HelmApplicationStatus struct {
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
	StatusTime *metav1.Time `json:"statusTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="application name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +genclient
// +genclient:nonNamespaced

// HelmApplication is the Schema for the helmapplications API
type HelmApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmApplicationSpec   `json:"spec,omitempty"`
	Status HelmApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelmApplicationList contains a list of HelmApplication
type HelmApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmApplication{}, &HelmApplicationList{})
}

func (in *HelmApplication) GetTrueName() string {
	if in == nil {
		return ""
	}
	return in.Spec.Name
}

func (in *HelmApplication) GetHelmRepoId() string {
	if in == nil {
		return ""
	}

	return getValue(in.Labels, constants.ChartRepoIdLabelKey)
}

func (in *HelmApplication) GetHelmApplicationId() string {
	if in == nil {
		return ""
	}
	return strings.TrimSuffix(in.Name, constants.HelmApplicationAppStoreSuffix)
}
func (in *HelmApplication) GetHelmCategoryId() string {
	if in == nil {
		return ""
	}

	return getValue(in.Labels, constants.CategoryIdLabelKey)
}

func (in *HelmApplication) GetWorkspace() string {
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

func (in *HelmApplication) GetCategoryId() string {
	if in == nil {
		return ""
	}
	return getValue(in.Labels, constants.CategoryIdLabelKey)
}
