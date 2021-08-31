/*
Copyright 2020 The KubeSphere authors.

Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha2

import (
	ants "kubesphere.io/monitoring-dashboard/api/v1alpha2/annotations"
	panels "kubesphere.io/monitoring-dashboard/api/v1alpha2/panels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	templatings "kubesphere.io/monitoring-dashboard/api/v1alpha2/templatings"
	time "kubesphere.io/monitoring-dashboard/api/v1alpha2/time"
)

// DashboardSpec defines the desired state of Dashboard
type DashboardSpec struct {
	ID              uint     `json:"id,omitempty" `
	UID             string   `json:"uid,omitempty" `
	Title           string   `json:"title,omitempty"`
	Description     string   `json:"description,omitempty"`
	Editable        bool     `json:"editable,omitempty"`
	SharedCrosshair bool     `json:"shared_crosshair,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	AutoRefresh     string   `json:"auto_refresh,omitempty"`
	Timezone        string   `json:"timezone,omitempty"`
	// Annotations
	Annotations []ants.Annotation `json:"annotations,omitempty"`
	// Time range
	Time   time.Time       `json:"time,omitempty"`
	Panels []*panels.Panel `json:"panels,omitempty"`
	// // Templating variables
	Templatings []templatings.TemplateVar `json:"templatings,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Dashboard is the Schema for the dashboards API
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DashboardSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// DashboardList contains a list of Dashboard
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:storageversion

// ClusterDashboard is the Schema for the culsterdashboards API
type ClusterDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DashboardSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterDashboardList contains a list of ClusterDashboard
type ClusterDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
	SchemeBuilder.Register(&ClusterDashboard{}, &ClusterDashboardList{})
}
