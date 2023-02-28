// Copyright (c) 2017-2018,2020-2021 Tigera, Inc. All rights reserved.

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

import (
	"github.com/projectcalico/api/pkg/lib/numorstring"
)

// PolicyType enumerates the possible values of the PolicySpec Types field.
type PolicyType string

const (
	PolicyTypeIngress PolicyType = "Ingress"
	PolicyTypeEgress  PolicyType = "Egress"
)

// A Rule encapsulates a set of match criteria and an action.  Both selector-based security Policy
// and security Profiles reference rules - separated out as a list of rules for both
// ingress and egress packet matching.
//
// Each positive match criteria has a negated version, prefixed with "Not". All the match
// criteria within a rule must be satisfied for a packet to match. A single rule can contain
// the positive and negative version of a match and both must be satisfied for the rule to match.
type Rule struct {
	Action Action `json:"action" validate:"action"`
	// IPVersion is an optional field that restricts the rule to only match a specific IP
	// version.
	IPVersion *int `json:"ipVersion,omitempty" validate:"omitempty,ipVersion"`
	// Protocol is an optional field that restricts the rule to only apply to traffic of
	// a specific IP protocol. Required if any of the EntityRules contain Ports
	// (because ports only apply to certain protocols).
	//
	// Must be one of these string values: "TCP", "UDP", "ICMP", "ICMPv6", "SCTP", "UDPLite"
	// or an integer in the range 1-255.
	Protocol *numorstring.Protocol `json:"protocol,omitempty" validate:"omitempty"`
	// ICMP is an optional field that restricts the rule to apply to a specific type and
	// code of ICMP traffic.  This should only be specified if the Protocol field is set to
	// "ICMP" or "ICMPv6".
	ICMP *ICMPFields `json:"icmp,omitempty" validate:"omitempty"`
	// NotProtocol is the negated version of the Protocol field.
	NotProtocol *numorstring.Protocol `json:"notProtocol,omitempty" validate:"omitempty"`
	// NotICMP is the negated version of the ICMP field.
	NotICMP *ICMPFields `json:"notICMP,omitempty" validate:"omitempty"`
	// Source contains the match criteria that apply to source entity.
	Source EntityRule `json:"source,omitempty" validate:"omitempty"`
	// Destination contains the match criteria that apply to destination entity.
	Destination EntityRule `json:"destination,omitempty" validate:"omitempty"`

	// HTTP contains match criteria that apply to HTTP requests.
	HTTP *HTTPMatch `json:"http,omitempty" validate:"omitempty"`

	// Metadata contains additional information for this rule
	Metadata *RuleMetadata `json:"metadata,omitempty" validate:"omitempty"`
}

// HTTPPath specifies an HTTP path to match. It may be either of the form:
// exact: <path>: which matches the path exactly or
// prefix: <path-prefix>: which matches the path prefix
type HTTPPath struct {
	Exact  string `json:"exact,omitempty" validate:"omitempty"`
	Prefix string `json:"prefix,omitempty" validate:"omitempty"`
}

// HTTPMatch is an optional field that apply only to HTTP requests
// The Methods and Path fields are joined with AND
type HTTPMatch struct {
	// Methods is an optional field that restricts the rule to apply only to HTTP requests that use one of the listed
	// HTTP Methods (e.g. GET, PUT, etc.)
	// Multiple methods are OR'd together.
	Methods []string `json:"methods,omitempty" validate:"omitempty"`
	// Paths is an optional field that restricts the rule to apply to HTTP requests that use one of the listed
	// HTTP Paths.
	// Multiple paths are OR'd together.
	// e.g:
	// - exact: /foo
	// - prefix: /bar
	// NOTE: Each entry may ONLY specify either a `exact` or a `prefix` match. The validator will check for it.
	Paths []HTTPPath `json:"paths,omitempty" validate:"omitempty"`
}

// ICMPFields defines structure for ICMP and NotICMP sub-struct for ICMP code and type
type ICMPFields struct {
	// Match on a specific ICMP type.  For example a value of 8 refers to ICMP Echo Request
	// (i.e. pings).
	Type *int `json:"type,omitempty" validate:"omitempty,gte=0,lte=254"`
	// Match on a specific ICMP code.  If specified, the Type value must also be specified.
	// This is a technical limitation imposed by the kernel's iptables firewall, which
	// Calico uses to enforce the rule.
	Code *int `json:"code,omitempty" validate:"omitempty,gte=0,lte=255"`
}

// An EntityRule is a sub-component of a Rule comprising the match criteria specific
// to a particular entity (that is either the source or destination).
//
// A source EntityRule matches the source endpoint and originating traffic.
// A destination EntityRule matches the destination endpoint and terminating traffic.
type EntityRule struct {
	// Nets is an optional field that restricts the rule to only apply to traffic that
	// originates from (or terminates at) IP addresses in any of the given subnets.
	Nets []string `json:"nets,omitempty" validate:"omitempty,dive,net"`

	// Selector is an optional field that contains a selector expression (see Policy for
	// sample syntax).  Only traffic that originates from (terminates at) endpoints matching
	// the selector will be matched.
	//
	// Note that: in addition to the negated version of the Selector (see NotSelector below), the
	// selector expression syntax itself supports negation.  The two types of negation are subtly
	// different. One negates the set of matched endpoints, the other negates the whole match:
	//
	//	Selector = "!has(my_label)" matches packets that are from other Calico-controlled
	// 	endpoints that do not have the label "my_label".
	//
	// 	NotSelector = "has(my_label)" matches packets that are not from Calico-controlled
	// 	endpoints that do have the label "my_label".
	//
	// The effect is that the latter will accept packets from non-Calico sources whereas the
	// former is limited to packets from Calico-controlled endpoints.
	Selector string `json:"selector,omitempty" validate:"omitempty,selector"`

	// NamespaceSelector is an optional field that contains a selector expression. Only traffic
	// that originates from (or terminates at) endpoints within the selected namespaces will be
	// matched. When both NamespaceSelector and another selector are defined on the same rule, then only
	// workload endpoints that are matched by both selectors will be selected by the rule.
	//
	// For NetworkPolicy, an empty NamespaceSelector implies that the Selector is limited to selecting
	// only workload endpoints in the same namespace as the NetworkPolicy.
	//
	// For NetworkPolicy, `global()` NamespaceSelector implies that the Selector is limited to selecting
	// only GlobalNetworkSet or HostEndpoint.
	//
	// For GlobalNetworkPolicy, an empty NamespaceSelector implies the Selector applies to workload
	// endpoints across all namespaces.
	NamespaceSelector string `json:"namespaceSelector,omitempty" validate:"omitempty,selector"`

	// Services is an optional field that contains options for matching Kubernetes Services.
	// If specified, only traffic that originates from or terminates at endpoints within the selected
	// service(s) will be matched, and only to/from each endpoint's port.
	//
	// Services cannot be specified on the same rule as Selector, NotSelector, NamespaceSelector, Nets,
	// NotNets or ServiceAccounts.
	//
	// Ports and NotPorts can only be specified with Services on ingress rules.
	Services *ServiceMatch `json:"services,omitempty" validate:"omitempty"`

	// Ports is an optional field that restricts the rule to only apply to traffic that has a
	// source (destination) port that matches one of these ranges/values. This value is a
	// list of integers or strings that represent ranges of ports.
	//
	// Since only some protocols have ports, if any ports are specified it requires the
	// Protocol match in the Rule to be set to "TCP" or "UDP".
	Ports []numorstring.Port `json:"ports,omitempty" validate:"omitempty,dive"`

	// NotNets is the negated version of the Nets field.
	NotNets []string `json:"notNets,omitempty" validate:"omitempty,dive,net"`

	// NotSelector is the negated version of the Selector field.  See Selector field for
	// subtleties with negated selectors.
	NotSelector string `json:"notSelector,omitempty" validate:"omitempty,selector"`

	// NotPorts is the negated version of the Ports field.
	// Since only some protocols have ports, if any ports are specified it requires the
	// Protocol match in the Rule to be set to "TCP" or "UDP".
	NotPorts []numorstring.Port `json:"notPorts,omitempty" validate:"omitempty,dive"`

	// ServiceAccounts is an optional field that restricts the rule to only apply to traffic that originates from (or
	// terminates at) a pod running as a matching service account.
	ServiceAccounts *ServiceAccountMatch `json:"serviceAccounts,omitempty" validate:"omitempty"`
}

type ServiceMatch struct {
	// Name specifies the name of a Kubernetes Service to match.
	Name string `json:"name,omitempty" validate:"omitempty,name"`

	// Namespace specifies the namespace of the given Service. If left empty, the rule
	// will match within this policy's namespace.
	Namespace string `json:"namespace,omitempty" validate:"omitempty,name"`
}

type ServiceAccountMatch struct {
	// Names is an optional field that restricts the rule to only apply to traffic that originates from (or terminates
	// at) a pod running as a service account whose name is in the list.
	Names []string `json:"names,omitempty" validate:"omitempty"`

	// Selector is an optional field that restricts the rule to only apply to traffic that originates from
	// (or terminates at) a pod running as a service account that matches the given label selector.
	// If both Names and Selector are specified then they are AND'ed.
	Selector string `json:"selector,omitempty" validate:"omitempty,selector"`
}

type Action string

const (
	Allow Action = "Allow"
	Deny         = "Deny"
	Log          = "Log"
	Pass         = "Pass"
)

type RuleMetadata struct {
	// Annotations is a set of key value pairs that give extra information about the rule
	Annotations map[string]string `json:"annotations,omitempty"`
}
