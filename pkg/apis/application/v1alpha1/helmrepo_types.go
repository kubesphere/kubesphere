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
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/constants"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	HELM_REPO_ID_LABEL                = "repo_id"
	HELM_APPLICATION_ID_LABEL         = "app_id"
	HELM_APPLICATION_VERSION_ID_LABEL = "version_id"

	WORKSPACE_FIELD = "workspace"
	NAME_FILED      = "name"
)

type HelmRepoCredential struct {
	//chart repository username
	Username string `json:"username,omitempty"`
	//chart repository password
	Password string `json:"password,omitempty"`
	//identify HTTPS client using this SSL certificate file
	CertFile string `json:"certFile,omitempty"`
	//identify HTTPS client using this SSL key file
	KeyFile string `json:"keyFile,omitempty"`
	//verify certificates of HTTPS-enabled servers using this CA bundle
	CAFile string `json:"caFile,omitempty"`
	//helm repo description
	Description string `json:"description,omitempty"`
	//skip tls certificate checks for the repository, default is ture
	InsecureSkipTLSVerify *bool `json:"insecureSkipTLSVerify,omitempty"`
}

// HelmRepoSpec defines the desired state of HelmRepo
type HelmRepoSpec struct {
	Creator   string `json:"creator,omitempty"`
	Name      string `json:"name"`
	Workspace string `json:"workspace"`
	//helm repo path
	Url string `json:"url"`
	//helm repo credential
	Credential HelmRepoCredential `json:"credential,omitempty"`
	//chart repo description
	Description string `json:"description,omitempty"`
	//sync period in seconds, no sync when SyncPeriod=0, the minimum SyncPeriod is 180s
	SyncPeriod int `json:"syncPeriod,omitempty"`
	// if ReSyncNow is true, then sync the repo now and set the value to false
	ReSyncNow bool `json:"reSyncNow,omitempty"`
}

type HelmRepoSyncState struct {
	//last sync state, valid state are: "failed", "success", and ""
	State string `json:"state,omitempty"`
	// A human readable message indicating details about why the repo is in this state.
	Message  string       `json:"message,omitempty"`
	SyncTime *metav1.Time `json:"syncTime"`
}

// HelmRepoStatus defines the observed state of HelmRepo

type HelmRepoStatus struct {
	//repo index, base64 encoded
	Data strfmt.Base64 `json:"data,omitempty"`
	//status last update time
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	SyncState []HelmRepoSyncState `json:"syncState,omitempty"`
	//total charts of the repo
	TotalChartVersions int `json:"totalChartVersions,omitempty"`
	//total chart versions of the repo
	TotalCharts int `json:"totalCharts,omitempty"`

	// if ReSyncNow is true, then sync the repo now and set the value to false
	ReSyncNow bool `json:"reSyncNow,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=helmrepos
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="name",type=string,JSONPath=`.spec.name`
// +genclient
// +genclient:nonNamespaced

// HelmRepo is the Schema for the helmrepoes API
type HelmRepo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmRepoSpec   `json:"spec,omitempty"`
	Status HelmRepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelmRepoList contains a list of HelmRepo
type HelmRepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRepo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRepo{}, &HelmRepoList{})
}

func (in *HelmRepo) GetTrueName() string {
	if in == nil {
		return ""
	}
	return in.Spec.Name
}

func (in *HelmRepo) GetHelmRepoId() string {
	if in == nil {
		return ""
	}
	return in.Name
}

func (in *HelmRepo) GetWorkspace() string {
	if in == nil {
		return ""
	}
	if l := in.Labels; l == nil {
		return ""
	} else {
		return l[constants.WorkspaceLabelKey]
	}
}
