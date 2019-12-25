package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FluentBit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FluentBitSpec   `json:"spec"`
	Status            FluentBitStatus `json:"status,omitempty"`
}

// FluentBitSpec holds the spec for the operator
type FluentBitSpec struct {
	Service  []Plugin `json:"service"`
	Input    []Plugin `json:"input"`
	Filter   []Plugin `json:"filter"`
	Output   []Plugin `json:"output"`
	Settings []Plugin `json:"settings"`
}

// FluentBitStatus holds the status info for the operator
type FluentBitStatus struct {
	// Fill me
}

// Plugin struct for fluent-bit plugins
type Plugin struct {
	Type       string      `json:"type" description:"output plugin type, eg. fluentbit-output-es"`
	Name       string      `json:"name" description:"output plugin name, eg. fluentbit-output-es"`
	Parameters []Parameter `json:"parameters" description:"output plugin configuration parameters"`
}

// Fluent-bit output plugins
type OutputPlugin struct {
	Plugin
	Id         string       `json:"id,omitempty" description:"output uuid"`
	Enable     bool         `json:"enable" description:"active status, one of true, false"`
	Updatetime *metav1.Time `json:"updatetime,omitempty" description:"last updatetime"`
}

// Parameter generic parameter type to handle values from different sources
type Parameter struct {
	Name      string     `json:"name" description:"configuration parameter key, eg. Name. refer to Fluent bit's Output Plugins Section for more configuration parameters."`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
	Value     string     `json:"value" description:"configuration parameter value, eg. es. refer to Fluent bit's Output Plugins Section for more configuration parameters."`
}

// ValueFrom generic type to determine value origin
type ValueFrom struct {
	SecretKeyRef KubernetesSecret `json:"secretKeyRef"`
}

// KubernetesSecret is a ValueFrom type
type KubernetesSecret struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
}

// FluentBitList auto generated by the sdk
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FluentBitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []FluentBit `json:"items"`
}
