// Package v1beta1 contains API Schema definitions for the iam v1beta1 API group
// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +groupName=iam.kubesphere.io
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1beta1"}

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
		&Category{},
		&CategoryList{},
		&Role{},
		&RoleList{},
		&ClusterRole{},
		&ClusterRoleList{},
		&WorkspaceRole{},
		&WorkspaceRoleList{},
		&GlobalRole{},
		&GlobalRoleList{},
		&RoleTemplate{},
		&RoleTemplateList{},
		&RoleBinding{},
		&RoleBindingList{},
		&ClusterRoleBinding{},
		&ClusterRoleBindingList{},
		&WorkspaceRoleBinding{},
		&WorkspaceRoleBindingList{},
		&GlobalRoleBinding{},
		&GlobalRoleBindingList{},
		&BuiltinRole{},
		&BuiltinRoleList{},
		&User{},
		&UserList{},
		&Group{},
		&GroupList{},
		&GroupBinding{},
		&GroupBindingList{},
		&LoginRecord{},
		&LoginRecordList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
