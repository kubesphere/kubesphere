/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */
package am

import (
	"fmt"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"net/http"
)

type AccessManagementInterface interface {
	GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error)
	GetWorkspaceRoleOfUser(username, workspace string) (*iamv1alpha2.WorkspaceRole, error)
	GetClusterRoleOfUser(username, cluster string) (*rbacv1.ClusterRole, error)
	GetNamespaceRoleOfUser(username, namespace string) (*rbacv1.Role, error)
	ListRoles(username string, query *query.Query) (*api.ListResult, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	ListWorkspaceRoles(query *query.Query) (*api.ListResult, error)
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)

	ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error)
	ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error)
	ListWorkspaceRoleBindings(username, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error)
	ListRoleBindings(username, namespace string) ([]*rbacv1.RoleBinding, error)

	GetRoleReferenceRules(roleRef rbacv1.RoleRef, bindingNamespace string) ([]rbacv1.PolicyRule, error)
}

type amOperator struct {
	ksinformer     ksinformers.SharedInformerFactory
	k8sinformer    k8sinformers.SharedInformerFactory
	resourceGetter *resourcev1alpha3.ResourceGetter
}

func NewAMOperator(factory informers.InformerFactory) AccessManagementInterface {
	return &amOperator{
		ksinformer:     factory.KubeSphereSharedInformerFactory(),
		k8sinformer:    factory.KubernetesSharedInformerFactory(),
		resourceGetter: resourcev1alpha3.NewResourceGetter(factory),
	}
}

func (am *amOperator) GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error) {

	roleBindings, err := am.ksinformer.Iam().V1alpha2().GlobalRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	userRoleBindings := make([]*iamv1alpha2.GlobalRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			userRoleBindings = append(userRoleBindings, roleBinding)
		}
	}

	if len(userRoleBindings) > 0 {
		role, err := am.ksinformer.Iam().V1alpha2().GlobalRoles().Lister().Get(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(roleBindings) > 1 {
			klog.Warningf("conflict global role binding, username: %s", username)
		}
		return role, nil
	}

	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:  iamv1alpha2.ResourceKindGlobalRoleBinding,
		},
		Message: fmt.Sprintf("global role binding not found for %s", username),
	}}

	return nil, err
}

func (am *amOperator) GetWorkspaceRoleOfUser(username, workspace string) (*iamv1alpha2.WorkspaceRole, error) {

	roleBindings, err := am.ksinformer.Iam().V1alpha2().WorkspaceRoleBindings().Lister().List(labels.SelectorFromValidatedSet(labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	userRoleBindings := make([]*iamv1alpha2.WorkspaceRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			userRoleBindings = append(userRoleBindings, roleBinding)
		}
	}

	if len(userRoleBindings) > 0 {
		role, err := am.ksinformer.Iam().V1alpha2().WorkspaceRoles().Lister().Get(userRoleBindings[0].RoleRef.Name)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if len(roleBindings) > 1 {
			klog.Warningf("conflict workspace role binding, username: %s", username)
		}

		return role, nil
	}

	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:  iamv1alpha2.ResourceKindWorkspaceRoleBinding,
		},
		Message: fmt.Sprintf("workspace role binding not found for %s", username),
	}}

	return nil, err
}

func (am *amOperator) GetNamespaceRoleOfUser(username, namespace string) (*rbacv1.Role, error) {
	roleBindings, err := am.k8sinformer.Rbac().V1().RoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	userRoleBindings := make([]*rbacv1.RoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			userRoleBindings = append(userRoleBindings, roleBinding)
		}
	}

	if len(userRoleBindings) > 0 {
		role, err := am.k8sinformer.Rbac().V1().Roles().Lister().Roles(namespace).Get(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(roleBindings) > 1 {
			klog.Warningf("conflict role binding, username: %s", username)
		}
		return role, nil
	}

	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: rbacv1.SchemeGroupVersion.Group,
			Kind:  "RoleBinding",
		},
		Message: fmt.Sprintf("role binding not found for %s in %s", username, namespace),
	}}

	return nil, err
}

// Get federated clusterrole of user if cluster is not empty, if cluster is empty means current cluster
func (am *amOperator) GetClusterRoleOfUser(username, cluster string) (*rbacv1.ClusterRole, error) {
	roleBindings, err := am.k8sinformer.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	userRoleBindings := make([]*rbacv1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			userRoleBindings = append(userRoleBindings, roleBinding)
		}
	}

	if len(userRoleBindings) > 0 {
		role, err := am.k8sinformer.Rbac().V1().ClusterRoles().Lister().Get(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(roleBindings) > 1 {
			klog.Warningf("conflict cluster role binding, username: %s", username)
		}
		return role, nil
	}
	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: rbacv1.SchemeGroupVersion.Group,
			Kind:  "ClusterRoleBinding",
		},
		Message: fmt.Sprintf("cluster role binding not found for %s in %s", username, cluster),
	}}

	return nil, err
}

func (am *amOperator) ListWorkspaceRoleBindings(username, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error) {
	roleBindings, err := am.ksinformer.Iam().V1alpha2().WorkspaceRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.WorkspaceRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		inSpecifiedWorkspace := workspace == "" || roleBinding.Labels[tenantv1alpha1.WorkspaceLabel] == workspace
		if contains(roleBinding.Subjects, username) && inSpecifiedWorkspace {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error) {

	roleBindings, err := am.k8sinformer.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*rbacv1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error) {
	roleBindings, err := am.ksinformer.Iam().V1alpha2().GlobalRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.GlobalRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListRoleBindings(username, namespace string) ([]*rbacv1.RoleBinding, error) {

	roleBindings, err := am.k8sinformer.Rbac().V1().RoleBindings().Lister().RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := make([]*rbacv1.RoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if contains(roleBinding.Subjects, username) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func contains(subjects []rbacv1.Subject, username string) bool {
	for _, subject := range subjects {
		if subject.Kind == rbacv1.UserKind && (username == "" || subject.Name == username) {
			return true
		}
	}
	return false
}

func (am *amOperator) ListRoles(namespace string, query *query.Query) (*api.ListResult, error) {
	return am.resourceGetter.List("roles", namespace, query)
}

func (am *amOperator) ListClusterRoles(query *query.Query) (*api.ListResult, error) {
	return am.resourceGetter.List("clusterroles", "", query)
}

func (am *amOperator) ListWorkspaceRoles(queryParam *query.Query) (*api.ListResult, error) {
	return am.resourceGetter.List(iamv1alpha2.ResourcesPluralWorkspaceRole, "", queryParam)
}

func (am *amOperator) ListGlobalRoles(query *query.Query) (*api.ListResult, error) {
	return am.resourceGetter.List(iamv1alpha2.ResourcesPluralGlobalRole, "", query)
}

// GetRoleReferenceRules attempts to resolve the RoleBinding or ClusterRoleBinding.
func (am *amOperator) GetRoleReferenceRules(roleRef rbacv1.RoleRef, bindingNamespace string) ([]rbacv1.PolicyRule, error) {
	switch roleRef.Kind {
	case "Role":
		role, err := am.k8sinformer.Rbac().V1().Roles().Lister().Roles(bindingNamespace).Get(roleRef.Name)
		if err != nil {
			return nil, err
		}
		return role.Rules, nil

	case "ClusterRole":
		clusterRole, err := am.k8sinformer.Rbac().V1().ClusterRoles().Lister().Get(roleRef.Name)
		if err != nil {
			return nil, err
		}
		return clusterRole.Rules, nil
	case iamv1alpha2.ResourceKindGlobalRole:
		globalRole, err := am.ksinformer.Iam().V1alpha2().GlobalRoles().Lister().Get(roleRef.Name)
		if err != nil {
			return nil, err
		}
		return globalRole.Rules, nil
	case iamv1alpha2.ResourceKindWorkspaceRole:
		workspaceRole, err := am.ksinformer.Iam().V1alpha2().WorkspaceRoles().Lister().Get(roleRef.Name)
		if err != nil {
			return nil, err
		}
		return workspaceRole.Rules, nil
	default:
		return nil, fmt.Errorf("unsupported role reference kind: %q", roleRef.Kind)
	}
}
