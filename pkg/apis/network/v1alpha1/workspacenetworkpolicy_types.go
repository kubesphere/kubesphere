/*
Copyright 2019 The KubeSphere authors.

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
	k8snetworkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkspaceNetworkPolicy     = "WorkspaceNetworkPolicy"
	ResourceSingularWorkspaceNetworkPolicy = "workspacenetworkpolicy"
	ResourcePluralWorkspaceNetworkPolicy   = "workspacenetworkpolicies"
)

// WorkspaceNetworkPolicySpec defines the desired state of WorkspaceNetworkPolicy
type WorkspaceNetworkPolicySpec struct {
	// Workspace specify the name of ws to apply this workspace network policy
	Workspace string `json:"workspace,omitempty"`
	// List of rule types that the WorkspaceNetworkPolicy relates to.
	// Valid options are Ingress, Egress, or Ingress,Egress.
	// If this field is not specified, it will default based on the existence of Ingress or Egress rules;
	// policies that contain an Egress section are assumed to affect Egress, and all policies
	// (whether or not they contain an Ingress section) are assumed to affect Ingress.
	// If you want to write an egress-only policy, you must explicitly specify policyTypes [ "Egress" ].
	// Likewise, if you want to write a policy that specifies that no egress is allowed,
	// you must specify a policyTypes value that include "Egress" (since such a policy would not include
	// an Egress section and would otherwise default to just [ "Ingress" ]).
	// +optional
	PolicyTypes []k8snetworkv1.PolicyType `json:"policyTypes,omitempty" protobuf:"bytes,4,rep,name=policyTypes,casttype=PolicyType"`
	// List of ingress rules to be applied to the selected pods. Traffic is allowed to
	// a pod if there are no NetworkPolicies selecting the pod
	// (and cluster policy otherwise allows the traffic), OR if the traffic source is
	// the pod's local node, OR if the traffic matches at least one ingress rule
	// across all of the NetworkPolicy objects whose podSelector matches the pod. If
	// this field is empty then this NetworkPolicy does not allow any traffic (and serves
	// solely to ensure that the pods it selects are isolated by default)
	// +optional
	Ingress []WorkspaceNetworkPolicyIngressRule `json:"ingress,omitempty" protobuf:"bytes,2,rep,name=ingress"`

	// List of egress rules to be applied to the selected pods. Outgoing traffic is
	// allowed if there are no NetworkPolicies selecting the pod (and cluster policy
	// otherwise allows the traffic), OR if the traffic matches at least one egress rule
	// across all of the NetworkPolicy objects whose podSelector matches the pod. If
	// this field is empty then this NetworkPolicy limits all outgoing traffic (and serves
	// solely to ensure that the pods it selects are isolated by default).
	// This field is beta-level in 1.8
	// +optional
	Egress []WorkspaceNetworkPolicyEgressRule `json:"egress,omitempty" protobuf:"bytes,3,rep,name=egress"`
}

// WorkspaceNetworkPolicyStatus defines the observed state of WorkspaceNetworkPolicy
type WorkspaceNetworkPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceNetworkPolicy is a set of network policies applied to the scope to workspace
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="networking",scope="Cluster",shortName="wsnp"
type WorkspaceNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceNetworkPolicySpec   `json:"spec,omitempty"`
	Status WorkspaceNetworkPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceNetworkPolicyList contains a list of WorkspaceNetworkPolicy
type WorkspaceNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceNetworkPolicy `json:"items"`
}

// WorkspaceNetworkPolicyIngressRule describes a particular set of traffic that is allowed to the pods
// matched by a WorkspaceNetworkPolicySpec's podSelector. The traffic must match both ports and from.
type WorkspaceNetworkPolicyIngressRule struct {
	// List of ports which should be made accessible on the pods selected for this
	// rule. Each item in this list is combined using a logical OR. If this field is
	// empty or missing, this rule matches all ports (traffic not restricted by port).
	// If this field is present and contains at least one item, then this rule allows
	// traffic only if the traffic matches at least one port in the list.
	// +optional
	Ports []k8snetworkv1.NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`

	// List of sources which should be able to access the pods selected for this rule.
	// Items in this list are combined using a logical OR operation. If this field is
	// empty or missing, this rule matches all sources (traffic not restricted by
	// source). If this field is present and contains at least on item, this rule
	// allows traffic only if the traffic matches at least one item in the from list.
	// +optional
	From []WorkspaceNetworkPolicyPeer `json:"from,omitempty" protobuf:"bytes,2,rep,name=from"`
}

// WorkspaceNetworkPolicyPeer describes a peer to allow traffic from. Only certain combinations of
// fields are allowed. It is same as 'NetworkPolicyPeer' in k8s but with an additional field 'WorkspaceSelector'
type WorkspaceNetworkPolicyPeer struct {
	k8snetworkv1.NetworkPolicyPeer `json:",inline"`
	WorkspaceSelector              *metav1.LabelSelector `json:"workspaceSelector,omitempty"`
}

// WorkspaceNetworkPolicyEgressRule describes a particular set of traffic that is allowed out of pods
// matched by a WorkspaceNetworkPolicySpec's podSelector. The traffic must match both ports and to.
type WorkspaceNetworkPolicyEgressRule struct {
	// List of ports which should be made accessible on the pods selected for this
	// rule. Each item in this list is combined using a logical OR. If this field is
	// empty or missing, this rule matches all ports (traffic not restricted by port).
	// If this field is present and contains at least one item, then this rule allows
	// traffic only if the traffic matches at least one port in the list.
	// +optional
	Ports []k8snetworkv1.NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`

	// List of sources which should be able to access the pods selected for this rule.
	// Items in this list are combined using a logical OR operation. If this field is
	// empty or missing, this rule matches all sources (traffic not restricted by
	// source). If this field is present and contains at least on item, this rule
	// allows traffic only if the traffic matches at least one item in the from list.
	// +optional
	To []WorkspaceNetworkPolicyPeer `json:"from,omitempty" protobuf:"bytes,2,rep,name=from"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceNetworkPolicy{}, &WorkspaceNetworkPolicyList{})
}
