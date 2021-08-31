/*
Copyright 2020 The KubeSphere authors.

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
	"kubesphere.io/monitoring-dashboard/api/v1alpha1/panels"
)

// DashboardSpec defines the desired state of Dashboard
type DashboardSpec struct {
	// Dashboard title
	Title string `json:"title,omitempty"`
	// Dashboard description
	Description string `json:"description,omitempty"`
	// Dashboard datasource
	DataSource string `json:"datasource,omitempty"`
	// Time range for display
	Time Time `json:"time,omitempty"`
	// Collection of panels. Panel is one of [Row](row.md), [Singlestat](#singlestat.md) or [Graph](graph.md)
	Panels []Panel `json:"panels,omitempty"`
	// Templating variables
	Templatings []Templating `json:"templating,omitempty"`
}

// Time ranges of the metrics for display
type Time struct {
	// Start time in the format of `^now([+-][0-9]+[smhdwMy])?$`, eg. `now-1M`.
	// It denotes the end time is set to the last month since now.
	From string `json:"from,omitempty"`
	// End time in the format of `^now([+-][0-9]+[smhdwMy])?$`, eg. `now-1M`.
	// It denotes the start time is set to the last month since now.
	To string `json:"to,omitempty"`
}

// Supported panel
type Panel struct {
	// panel metadata
	PanelMeta `json:",inline"`

	// A collection of queries
	// Only for panels with `graph` or `singlestat` type
	Targets []panels.Target `json:"targets,omitempty"`

	// The panel row
	Row *panels.Row `json:",inline"`
	// The panel graph
	Graph *panels.Graph `json:",inline"`
	// The panel singlestat
	SingleStat *panels.SingleStat `json:",inline"`
}

type PanelType string

const (
	PanelRow        PanelType = "row"
	PanelGraph      PanelType = "graph"
	PanelSingleStat PanelType = "singlestat"
)

type PanelMeta struct {
	// Name of the panel
	Title string `json:"title,omitempty"`
	// Panel ID
	Id int64 `json:"id,omitempty"`
	// Panel Type, one of `row`, `graph`, `singlestat`
	Type PanelType `json:"type"`
}

// Templating defines a variable, which can be used as a placeholder in query
type Templating struct {
	// Variable name
	Name string `json:"name,omitempty"`
	// Set variable values to be the return result of the query
	Query string `json:"query,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
