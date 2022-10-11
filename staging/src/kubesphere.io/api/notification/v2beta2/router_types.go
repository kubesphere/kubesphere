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

package v2beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Channel struct {
	Tenant string `json:"tenant"`
	// Receiver type, known values are dingtalk, email, feishu, slack, sms, pushover, webhook, wechat.
	Type []string `json:"type,omitempty"`
}

type ReceiverSelector struct {
	Name      []string              `json:"name,omitempty"`
	RegexName string                `json:"regexName,omitempty"`
	Selector  *metav1.LabelSelector `json:"selector,omitempty"`
	Channels  []Channel             `json:"channels,omitempty"`
	// Receiver type, known values are dingtalk, email, feishu, slack, sms, pushover, webhook, wechat.
	Type string `json:"type,omitempty"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	// whether the router is enabled
	Enabled       *bool                 `json:"enabled,omitempty"`
	AlertSelector *metav1.LabelSelector `json:"alertSelector"`
	// Receivers which need to receive the matched alert.
	Receivers ReceiverSelector `json:"receivers"`
}

// RouterStatus defines the observed state of Router
type RouterStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories=notification-manager
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +genclient:nonNamespaced
// Router is the Schema for the router API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec   `json:"spec,omitempty"`
	Status RouterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouterList contains a list of Router
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Router{}, &RouterList{})
}
