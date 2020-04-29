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

package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

type GenericClusterReference struct {
	Name string `json:"name"`
}

type GenericPlacementFields struct {
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

func UnmarshalGenericPlacement(obj *unstructured.Unstructured) (*GenericPlacement, error) {
	placement := &GenericPlacement{}
	err := UnstructuredToInterface(obj, placement)
	if err != nil {
		return nil, err
	}
	return placement, nil
}

func (p *GenericPlacement) ClusterNames() []string {
	if p.Spec.Placement.Clusters == nil {
		return nil
	}
	clusterNames := []string{}
	for _, cluster := range p.Spec.Placement.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}
	return clusterNames
}

func (p *GenericPlacement) ClusterSelector() (labels.Selector, error) {
	return metav1.LabelSelectorAsSelector(p.Spec.Placement.ClusterSelector)
}

func GetClusterNames(obj *unstructured.Unstructured) ([]string, error) {
	placement, err := UnmarshalGenericPlacement(obj)
	if err != nil {
		return nil, err
	}
	return placement.ClusterNames(), nil
}

func SetClusterNames(obj *unstructured.Unstructured, clusterNames []string) error {
	var clusters []interface{}
	if clusterNames != nil {
		clusters = []interface{}{}
		for _, clusterName := range clusterNames {
			clusters = append(clusters, map[string]interface{}{
				NameField: clusterName,
			})
		}
	}
	return unstructured.SetNestedSlice(obj.Object, clusters, SpecField, PlacementField, ClustersField)
}
