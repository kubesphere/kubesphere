// Package v2 contains API Schema definitions for the application v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=application.kubesphere.io
package v2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "application.kubesphere.io", Version: "v2"}
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(
		&Category{},
		&CategoryList{},
		&Application{},
		&ApplicationList{},
		&ApplicationVersion{},
		&ApplicationVersionList{},
		&ApplicationRelease{},
		&ApplicationReleaseList{},
		&Repo{},
		&RepoList{},
	)
}
