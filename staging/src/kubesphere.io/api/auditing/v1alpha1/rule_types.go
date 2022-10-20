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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PolicyRule struct {
	// Rule name
	Name string `json:"name,omitempty" protobuf:"bytes,8,opt,name=name"`
	// Rule type, rule, macro,list,alias
	Type string `json:"type,omitempty" protobuf:"bytes,8,opt,name=type"`
	// Rule describe
	Desc string `json:"desc,omitempty" protobuf:"bytes,8,opt,name=desc"`
	// Rule condition
	// This effective When the rule type is rule
	Condition string `json:"condition,omitempty" protobuf:"bytes,8,opt,name=condition"`
	// This effective When the rule type is macro
	Macro string `json:"macro,omitempty" protobuf:"bytes,8,opt,name=macro"`
	// This effective When the rule type is alias
	Alias string `json:"alias,omitempty" protobuf:"bytes,8,opt,name=alias"`
	// This effective When the rule type is list
	List []string `json:"list,omitempty" protobuf:"bytes,8,opt,name=list"`
	// Is the rule enable
	Enable bool `json:"enable" protobuf:"bytes,8,opt,name=enable"`
	// The output formater of message which send to user
	Output string `json:"output,omitempty" protobuf:"bytes,8,opt,name=output"`
	// Rule priority, DEBUG, INFO, WARNING
	Priority string `json:"priority,omitempty" protobuf:"bytes,8,opt,name=priority"`
}

// AuditRuleSpec defines the desired state of Rule
type RuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PolicyRules []PolicyRule `json:"rules,omitempty" protobuf:"bytes,8,opt,name=rules"`
}

// AuditRuleStatus defines the observed state of Rule
type RuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +kubebuilder:object:root=true

// Rule is the Schema for the rules API
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleSpec   `json:"spec,omitempty"`
	Status RuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AuditRuleList contains a list of Rule
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rule{}, &RuleList{})
}
