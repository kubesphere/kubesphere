// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ServiceMonitorsKind   = "ServiceMonitor"
	ServiceMonitorName    = "servicemonitors"
	ServiceMonitorKindKey = "servicemonitor"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="smon"

// ServiceMonitor defines monitoring for a set of services.
type ServiceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Service selection for target discovery by
	// Prometheus.
	Spec ServiceMonitorSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ServiceMonitor) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ServiceMonitorSpec contains specification parameters for a ServiceMonitor.
// +k8s:openapi-gen=true
type ServiceMonitorSpec struct {
	// JobLabel selects the label from the associated Kubernetes service which will be used as the `job` label for all metrics.
	//
	// For example:
	// If in `ServiceMonitor.spec.jobLabel: foo` and in `Service.metadata.labels.foo: bar`,
	// then the `job="bar"` label is added to all metrics.
	//
	// If the value of this field is empty or if the label doesn't exist for the given Service, the `job` label of the metrics defaults to the name of the Kubernetes Service.
	JobLabel string `json:"jobLabel,omitempty"`
	// TargetLabels transfers labels from the Kubernetes `Service` onto the created metrics.
	TargetLabels []string `json:"targetLabels,omitempty"`
	// PodTargetLabels transfers labels on the Kubernetes `Pod` onto the created metrics.
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
	// A list of endpoints allowed as part of this ServiceMonitor.
	Endpoints []Endpoint `json:"endpoints"`
	// Selector to select Endpoints objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the Kubernetes Endpoints objects are discovered from.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	SampleLimit uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	TargetLimit uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelLimit uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelNameLengthLimit uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelValueLengthLimit uint64 `json:"labelValueLengthLimit,omitempty"`
	// Attaches node metadata to discovered targets.
	// Requires Prometheus v2.37.0 and above.
	AttachMetadata *AttachMetadata `json:"attachMetadata,omitempty"`
}

// ServiceMonitorList is a list of ServiceMonitors.
// +k8s:openapi-gen=true
type ServiceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of ServiceMonitors
	Items []*ServiceMonitor `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ServiceMonitorList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
