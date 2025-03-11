// Package v1beta1 contains API Schema definitions for the tenant v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=tenant.kubesphere.io

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "tenant.kubesphere.io", Version: "v1beta1"}

	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	SchemeBuilder.Register(&Workspace{},
		&WorkspaceList{},
		&WorkspaceTemplate{},
		&WorkspaceTemplateList{})
}
