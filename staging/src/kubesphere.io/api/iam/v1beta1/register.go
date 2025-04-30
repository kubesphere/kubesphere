/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "iam.kubesphere.io"

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1beta1"}
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

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
