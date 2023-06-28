/*

 Copyright 2020 The KubeSphere Authors.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type GenericClusterReference struct {
	Name string `json:"name"`
}

type GenericPlacementFields struct {
	// +listType=map
	// +listMapKey=name
	Clusters        []GenericClusterReference `json:"clusters,omitempty"`
	ClusterSelector *metav1.LabelSelector     `json:"clusterSelector,omitempty"`
}
type GenericPlacementSpec struct {
	Placement GenericPlacementFields `json:"placement,omitempty"`
}

type GenericPlacement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GenericPlacementSpec `json:"spec,omitempty"`
}

type ClusterOverride struct {
	Op   string `json:"op,omitempty"`
	Path string `json:"path"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Value runtime.RawExtension `json:"value,omitempty"`
}

type GenericOverrideItem struct {
	ClusterName      string            `json:"clusterName"`
	ClusterOverrides []ClusterOverride `json:"clusterOverrides,omitempty"`
}

type GenericOverrideSpec struct {
	Overrides []GenericOverrideItem `json:"overrides,omitempty"`
}

type GenericOverride struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *GenericOverrideSpec `json:"spec,omitempty"`
}

type ConditionType string

type AggregateReason string

type PropagationStatus string

type GenericClusterStatus struct {
	Name   string            `json:"name"`
	Status PropagationStatus `json:"status,omitempty"`
}

type GenericCondition struct {
	// Type of cluster condition
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time reconciliation resulted in an error or the last time a
	// change was propagated to member clusters.
	// +optional
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason AggregateReason `json:"reason,omitempty"`
}

type GenericFederatedStatus struct {
	ObservedGeneration int64                  `json:"observedGeneration,omitempty"`
	Conditions         []*GenericCondition    `json:"conditions,omitempty"`
	Clusters           []GenericClusterStatus `json:"clusters,omitempty"`
}

type GenericFederatedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}
