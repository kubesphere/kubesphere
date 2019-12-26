/*
Copyright 2019 The Kubesphere Authors.

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

const (
	ResourceKindS2iBuilderTemplate     = "S2iBuilderTemplate"
	ResourceSingularS2iBuilderTemplate = "s2ibuildertemplate"
	ResourcePluralS2iBuilderTemplate   = "s2ibuildertemplates"
)

type Parameter struct {
	Description  string   `json:"description,omitempty"`
	Key          string   `json:"key,omitempty"`
	Type         string   `json:"type,omitempty"`
	OptValues    []string `json:"optValues,omitempty"`
	Required     bool     `json:"required,omitempty"`
	DefaultValue string   `json:"defaultValue,omitempty"`
	Value        string   `json:"value,omitempty"`
}

func (p *Parameter) ToEnvonment() *EnvironmentSpec {
	var v string
	if p.Value == "" && p.DefaultValue != "" {
		v = p.DefaultValue
	} else if p.Value != "" {
		v = p.Value
	} else {
		return nil
	}
	return &EnvironmentSpec{
		Name:  p.Key,
		Value: v,
	}
}

// S2iBuilderTemplateSpec defines the desired state of S2iBuilderTemplate
type S2iBuilderTemplateSpec struct {
	//DefaultBaseImage is the image that will be used by default
	DefaultBaseImage string `json:"defaultBaseImage,omitempty"`
	//Images are the images this template will use.
	ContainerInfo []ContainerInfo `json:"containerInfo,omitempty"`
	//CodeFramework means which language this template is designed for and which framework is using if has framework. Like Java, NodeJS etc
	CodeFramework CodeFramework `json:"codeFramework,omitempty"`
	// Parameters is a set of environment variables to be passed to the image.
	Parameters []Parameter `json:"environment,omitempty"`
	// Version of template
	Version string `json:"version,omitempty"`
	// Description illustrate the purpose of this template
	Description string `json:"description,omitempty"`
	// IconPath is used for frontend display
	IconPath string `json:"iconPath,omitempty"`
}

type ContainerInfo struct {
	//BaseImage are the images this template will use.
	BuilderImage     string       `json:"builderImage,omitempty"`
	RuntimeImage     string       `json:"runtimeImage,omitempty"`
	RuntimeArtifacts []VolumeSpec `json:"runtimeArtifacts,omitempty"`
	// BuildVolumes specifies a list of volumes to mount to container running the
	// build.
	BuildVolumes []string `json:"buildVolumes,omitempty"`
}

// S2iBuilderTemplateStatus defines the observed state of S2iBuilderTemplate
type S2iBuilderTemplateStatus struct {
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBuilderTemplate is the Schema for the s2ibuildertemplates API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Framework",type="string",JSONPath=".spec.codeFramework"
// +kubebuilder:printcolumn:name="DefaultBaseImage",type="string",JSONPath=".spec.defaultBaseImage"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:resource:categories="devops",scope="Cluster",shortName="s2ibt"
type S2iBuilderTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S2iBuilderTemplateSpec   `json:"spec,omitempty"`
	Status S2iBuilderTemplateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBuilderTemplateList contains a list of S2iBuilderTemplate
type S2iBuilderTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S2iBuilderTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S2iBuilderTemplate{}, &S2iBuilderTemplateList{})
}
