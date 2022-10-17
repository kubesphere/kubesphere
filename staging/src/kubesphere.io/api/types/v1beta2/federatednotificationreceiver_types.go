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
	ResourcePluralFederatedNotificationReceiver   = "federatednotificationreceivers"
	ResourceSingularFederatedNotificationReceiver = "federatednotificationreceiver"
	FederatedNotificationReceiverKind             = "FederatedNotificationReceiver"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

type FederatedNotificationReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedNotificationReceiverSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedNotificationReceiverSpec struct {
	Template  NotificationReceiverTemplate `json:"template"`
	Placement GenericPlacementFields       `json:"placement"`
	Overrides []GenericOverrideItem        `json:"overrides,omitempty"`
}

type NotificationReceiverTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v2beta2.ReceiverSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// FederatedNotificationReceiverList contains a list of federatednotificationreceiverlists
type FederatedNotificationReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedNotificationReceiver `json:"items"`
}

// ConvertTo converts this Config to the Hub version (v1beta1).
func (src *FederatedNotificationReceiver) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v1beta1.FederatedNotificationReceiver)
	dst.ObjectMeta = src.ObjectMeta

	srcReceiver := v2beta2.Receiver{
		Spec: src.Spec.Template.Spec,
	}
	dstReceiver := &v2beta1.Receiver{}
	if err := srcReceiver.ConvertTo(dstReceiver); err != nil {
		return err
	}

	dst.Spec = v1beta1.FederatedNotificationReceiverSpec{
		Template: v1beta1.NotificationReceiverTemplate{
			Spec: dstReceiver.Spec,
		},
		Placement: convertPlacementTo(src.Spec.Placement),
		Overrides: convertOverridesTo(src.Spec.Overrides),
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *FederatedNotificationReceiver) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v1beta1.FederatedNotificationReceiver)
	dst.ObjectMeta = src.ObjectMeta

	srcReceiver := v2beta1.Receiver{
		Spec: src.Spec.Template.Spec,
	}
	dstReceiver := &v2beta2.Receiver{}
	if err := dstReceiver.ConvertFrom(&srcReceiver); err != nil {
		return err
	}

	dst.Spec = FederatedNotificationReceiverSpec{
		Template: NotificationReceiverTemplate{
			Spec: dstReceiver.Spec,
		},
		Placement: convertPlacementFrom(src.Spec.Placement),
		Overrides: convertOverridesFrom(src.Spec.Overrides),
	}

	return nil
}
