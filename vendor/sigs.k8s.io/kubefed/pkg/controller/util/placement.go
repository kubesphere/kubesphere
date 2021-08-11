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
	"k8s.io/apimachinery/pkg/util/sets"

	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
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

func SetClusterSelector(obj *unstructured.Unstructured, clusterSelector map[string]string) error {
	return unstructured.SetNestedStringMap(obj.Object, clusterSelector, SpecField, PlacementField, ClusterSelectorField, MatchLabelsField)
}

// ComputeNamespacedPlacement determines placement for namespaced
// federated resources (e.g. FederatedConfigMap).
//
// If KubeFed is deployed cluster-wide, placement is the intersection
// of the placement for the federated resource and the placement of
// the federated namespace containing the resource.
//
// If KubeFed is limited to a single namespace, placement is
// determined as the intersection of resource and namespace placement
// if namespace placement exists.  If namespace placement does not
// exist, resource placement will be used verbatim.  This is possible
// because the single namespace by definition must exist on member
// clusters, so namespace placement becomes a mechanism for limiting
// rather than allowing propagation.
func ComputeNamespacedPlacement(resource, namespace *unstructured.Unstructured, clusters []*fedv1b1.KubeFedCluster, limitedScope bool, selectorOnly bool) (selectedClusters sets.String, err error) {
	resourceClusters, err := ComputePlacement(resource, clusters, selectorOnly)
	if err != nil {
		return nil, err
	}

	if namespace == nil {
		if limitedScope {
			// Use the resource placement verbatim if no federated
			// namespace is present and KubeFed is targeting a
			// single namespace.
			return resourceClusters, nil
		}
		// Resource should not exist in any member clusters.
		return sets.String{}, nil
	}

	namespaceClusters, err := ComputePlacement(namespace, clusters, selectorOnly)
	if err != nil {
		return nil, err
	}

	// If both namespace and resource placement exist, the desired
	// list of clusters is their intersection.
	return resourceClusters.Intersection(namespaceClusters), nil
}

// ComputePlacement determines the selected clusters for a federated
// resource.
func ComputePlacement(resource *unstructured.Unstructured, clusters []*fedv1b1.KubeFedCluster, selectorOnly bool) (selectedClusters sets.String, err error) {
	selectedNames, err := selectedClusterNames(resource, clusters, selectorOnly)
	if err != nil {
		return nil, err
	}
	clusterNames := getClusterNames(clusters)
	return clusterNames.Intersection(selectedNames), nil
}

func selectedClusterNames(resource *unstructured.Unstructured, clusters []*fedv1b1.KubeFedCluster, selectorOnly bool) (sets.String, error) {
	placement, err := UnmarshalGenericPlacement(resource)
	if err != nil {
		return nil, err
	}

	selectedNames := sets.String{}
	clusterNames := placement.ClusterNames()
	// Only use selector if clusters are nil. An empty list of
	// clusters implies no clusters are selected.
	if selectorOnly || clusterNames == nil {
		selector, err := placement.ClusterSelector()
		if err != nil {
			return nil, err
		}
		for _, cluster := range clusters {
			if selector.Matches(labels.Set(cluster.Labels)) {
				selectedNames.Insert(cluster.Name)
			}
		}
	} else {
		for _, clusterName := range clusterNames {
			selectedNames.Insert(clusterName)
		}
	}

	return selectedNames, nil
}

func getClusterNames(clusters []*fedv1b1.KubeFedCluster) sets.String {
	clusterNames := sets.String{}
	for _, cluster := range clusters {
		clusterNames.Insert(cluster.Name)
	}
	return clusterNames
}
