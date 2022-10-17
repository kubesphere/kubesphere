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

package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/api/notification/v2beta1"
	"kubesphere.io/api/notification/v2beta2"
	"kubesphere.io/api/types/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	ResourcePluralFederatedNotificationConfig   = "federatednotificationconfigs"
	ResourceSingularFederatedNotificationConfig = "federatednotificationconfig"
	FederatedNotificationConfigKind             = "FederatedNotificationConfig"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

type FederatedNotificationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedNotificationConfigSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedNotificationConfigSpec struct {
	Template  NotificationConfigTemplate `json:"template"`
	Placement GenericPlacementFields     `json:"placement"`
	Overrides []GenericOverrideItem      `json:"overrides,omitempty"`
}

type NotificationConfigTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v2beta2.ConfigSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// FederatedNotificationConfigList contains a list of federatednotificationconfiglists
type FederatedNotificationConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedNotificationConfig `json:"items"`
}

// ConvertTo converts this Config to the Hub version (v1beta1).
func (src *FederatedNotificationConfig) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v1beta1.FederatedNotificationConfig)
	dst.ObjectMeta = src.ObjectMeta

	srcConfig := v2beta2.Config{
		Spec: src.Spec.Template.Spec,
	}
	dstConfig := &v2beta1.Config{}
	if err := srcConfig.ConvertTo(dstConfig); err != nil {
		return err
	}

	dst.Spec = v1beta1.FederatedNotificationConfigSpec{
		Template: v1beta1.NotificationConfigTemplate{
			Spec: dstConfig.Spec,
		},
		Placement: convertPlacementTo(src.Spec.Placement),
		Overrides: convertOverridesTo(src.Spec.Overrides),
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *FederatedNotificationConfig) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v1beta1.FederatedNotificationConfig)
	dst.ObjectMeta = src.ObjectMeta

	srcConfig := v2beta1.Config{
		Spec: src.Spec.Template.Spec,
	}
	dstConfig := &v2beta2.Config{}
	if err := dstConfig.ConvertFrom(&srcConfig); err != nil {
		return err
	}

	dst.Spec = FederatedNotificationConfigSpec{
		Template: NotificationConfigTemplate{
			Spec: dstConfig.Spec,
		},
		Placement: convertPlacementFrom(src.Spec.Placement),
		Overrides: convertOverridesFrom(src.Spec.Overrides),
	}

	return nil
}

func convertPlacementTo(placement GenericPlacementFields) v1beta1.GenericPlacementFields {

	var clusters []v1beta1.GenericClusterReference
	for _, cluster := range placement.Clusters {
		clusters = append(clusters, v1beta1.GenericClusterReference{
			Name: cluster.Name,
		})
	}

	return v1beta1.GenericPlacementFields{
		Clusters:        clusters,
		ClusterSelector: placement.ClusterSelector,
	}
}

func convertPlacementFrom(placement v1beta1.GenericPlacementFields) GenericPlacementFields {

	var clusters []GenericClusterReference
	for _, cluster := range placement.Clusters {
		clusters = append(clusters, GenericClusterReference{
			Name: cluster.Name,
		})
	}

	return GenericPlacementFields{
		Clusters:        clusters,
		ClusterSelector: placement.ClusterSelector,
	}
}

func convertOverridesTo(src []GenericOverrideItem) []v1beta1.GenericOverrideItem {

	var dst []v1beta1.GenericOverrideItem
	for _, item := range src {
		overrideItem := v1beta1.GenericOverrideItem{ClusterName: item.ClusterName}
		for _, clusterOverride := range item.ClusterOverrides {
			overrideItem.ClusterOverrides = append(overrideItem.ClusterOverrides, v1beta1.ClusterOverride{
				Op:    clusterOverride.Op,
				Path:  clusterOverride.Path,
				Value: clusterOverride.Value,
			})
		}
		dst = append(dst, overrideItem)
	}

	return dst
}

func convertOverridesFrom(src []v1beta1.GenericOverrideItem) []GenericOverrideItem {

	var dst []GenericOverrideItem
	for _, item := range src {
		overrideItem := GenericOverrideItem{ClusterName: item.ClusterName}
		for _, clusterOverride := range item.ClusterOverrides {
			overrideItem.ClusterOverrides = append(overrideItem.ClusterOverrides, ClusterOverride{
				Op:    clusterOverride.Op,
				Path:  clusterOverride.Path,
				Value: clusterOverride.Value,
			})
		}
		dst = append(dst, overrideItem)
	}

	return dst
}
