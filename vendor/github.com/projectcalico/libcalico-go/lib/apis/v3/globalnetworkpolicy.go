// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	KindGlobalNetworkPolicy     = "GlobalNetworkPolicy"
	KindGlobalNetworkPolicyList = "GlobalNetworkPolicyList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalNetworkPolicy contains information about a security Policy resource.  This contains a set of
// security rules to apply.  Security policies allow a selector-based security model which can override
// the security profiles directly referenced by an endpoint.
//
// Each policy must do one of the following:
//
//  	- Match the packet and apply an “allow” action; this immediately accepts the packet, skipping
//        all further policies and profiles. This is not recommended in general, because it prevents
//        further policy from being executed.
// 	- Match the packet and apply a “deny” action; this drops the packet immediately, skipping all
//        further policy and profiles.
// 	- Fail to match the packet; in which case the packet proceeds to the next policy. If there
// 	  are no more policies then the packet is dropped.
//
// Calico implements the security policy for each endpoint individually and only the policies that
// have matching selectors are implemented. This ensures that the number of rules that actually need
// to be inserted into the kernel is proportional to the number of local endpoints rather than the
// total amount of policy.
//
// GlobalNetworkPolicy is globally-scoped (i.e. not Namespaced).
type GlobalNetworkPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the Policy.
	Spec GlobalNetworkPolicySpec `json:"spec,omitempty"`
}

type GlobalNetworkPolicySpec struct {
	// Order is an optional field that specifies the order in which the policy is applied.
	// Policies with higher "order" are applied after those with lower
	// order.  If the order is omitted, it may be considered to be "infinite" - i.e. the
	// policy will be applied last.  Policies with identical order will be applied in
	// alphanumerical order based on the Policy "Name".
	Order *float64 `json:"order,omitempty"`
	// The ordered set of ingress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	Ingress []Rule `json:"ingress,omitempty" validate:"omitempty,dive"`
	// The ordered set of egress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	Egress []Rule `json:"egress,omitempty" validate:"omitempty,dive"`
	// The selector is an expression used to pick pick out the endpoints that the policy should
	// be applied to.
	//
	// Selector expressions follow this syntax:
	//
	// 	label == "string_literal"  ->  comparison, e.g. my_label == "foo bar"
	// 	label != "string_literal"   ->  not equal; also matches if label is not present
	// 	label in { "a", "b", "c", ... }  ->  true if the value of label X is one of "a", "b", "c"
	// 	label not in { "a", "b", "c", ... }  ->  true if the value of label X is not one of "a", "b", "c"
	// 	has(label_name)  -> True if that label is present
	// 	! expr -> negation of expr
	// 	expr && expr  -> Short-circuit and
	// 	expr || expr  -> Short-circuit or
	// 	( expr ) -> parens for grouping
	// 	all() or the empty selector -> matches all endpoints.
	//
	// Label names are allowed to contain alphanumerics, -, _ and /. String literals are more permissive
	// but they do not support escape characters.
	//
	// Examples (with made-up labels):
	//
	// 	type == "webserver" && deployment == "prod"
	// 	type in {"frontend", "backend"}
	// 	deployment != "dev"
	// 	! has(label_name)
	Selector string `json:"selector,omitempty" validate:"selector"`
	// Types indicates whether this policy applies to ingress, or to egress, or to both.  When
	// not explicitly specified (and so the value on creation is empty or nil), Calico defaults
	// Types according to what Ingress and Egress rules are present in the policy.  The
	// default is:
	//
	// - [ PolicyTypeIngress ], if there are no Egress rules (including the case where there are
	//   also no Ingress rules)
	//
	// - [ PolicyTypeEgress ], if there are Egress rules but no Ingress rules
	//
	// - [ PolicyTypeIngress, PolicyTypeEgress ], if there are both Ingress and Egress rules.
	//
	// When the policy is read back again, Types will always be one of these values, never empty
	// or nil.
	Types []PolicyType `json:"types,omitempty" validate:"omitempty,dive,policyType"`

	// DoNotTrack indicates whether packets matched by the rules in this policy should go through
	// the data plane's connection tracking, such as Linux conntrack.  If True, the rules in
	// this policy are applied before any data plane connection tracking, and packets allowed by
	// this policy are marked as not to be tracked.
	DoNotTrack bool `json:"doNotTrack,omitempty"`
	// PreDNAT indicates to apply the rules in this policy before any DNAT.
	PreDNAT bool `json:"preDNAT,omitempty"`
	// ApplyOnForward indicates to apply the rules in this policy on forward traffic.
	ApplyOnForward bool `json:"applyOnForward,omitempty"`

	// ServiceAccountSelector is an optional field for an expression used to select a pod based on service accounts.
	ServiceAccountSelector string `json:"serviceAccountSelector,omitempty" validate:"selector"`

	// NamespaceSelector is an optional field for an expression used to select a pod based on namespaces.
	NamespaceSelector string `json:"namespaceSelector,omitempty" validate"selector"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalNetworkPolicyList contains a list of GlobalNetworkPolicy resources.
type GlobalNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []GlobalNetworkPolicy `json:"items"`
}

// NewGlobalNetworkPolicy creates a new (zeroed) GlobalNetworkPolicy struct with the TypeMetadata initialised to the current
// version.
func NewGlobalNetworkPolicy() *GlobalNetworkPolicy {
	return &GlobalNetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindGlobalNetworkPolicy,
			APIVersion: GroupVersionCurrent,
		},
	}
}

// NewGlobalNetworkPolicyList creates a new (zeroed) GlobalNetworkPolicyList struct with the TypeMetadata initialised to the current
// version.
func NewGlobalNetworkPolicyList() *GlobalNetworkPolicyList {
	return &GlobalNetworkPolicyList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindGlobalNetworkPolicyList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
