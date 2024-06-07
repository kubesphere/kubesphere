// NOTE: Boilerplate only. Ignore this file.

// Package v1alpha2 contains API Schema definitions for the iam v1alpha2 API group
// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +groupName=iam.kubesphere.io
package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme is required by pkg/client/...
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&User{},
		&UserList{},
		&LoginRecord{},
		&LoginRecordList{},
		&GlobalRole{},
		&GlobalRoleList{},
		&GlobalRoleBinding{},
		&GlobalRoleBindingList{},
		&WorkspaceRole{},
		&WorkspaceRoleList{},
		&WorkspaceRoleBinding{},
		&WorkspaceRoleBindingList{},
		&Group{},
		&GroupList{},
		&GroupBinding{},
		&GroupBindingList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
