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
package am

import (
	"encoding/json"
	"fmt"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"net/http"
)

type AccessManagementInterface interface {
	GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error)
	GetWorkspaceRoleOfUser(username, workspace string) (*iamv1alpha2.WorkspaceRole, error)
	GetClusterRoleOfUser(username string) (*rbacv1.ClusterRole, error)
	GetNamespaceRoleOfUser(username, namespace string) (*rbacv1.Role, error)
	ListRoles(username string, query *query.Query) (*api.ListResult, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	ListWorkspaceRoles(query *query.Query) (*api.ListResult, error)
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)

	ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error)
	ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error)
	ListWorkspaceRoleBindings(username, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error)
	ListRoleBindings(username, namespace string) ([]*rbacv1.RoleBinding, error)

	GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (string, []rbacv1.PolicyRule, error)
	GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error)
	GetWorkspaceRole(workspace string, name string) (*iamv1alpha2.WorkspaceRole, error)
	CreateOrUpdateGlobalRoleBinding(username string, globalRole string) error
	CreateOrUpdateWorkspaceRole(workspace string, workspaceRole *iamv1alpha2.WorkspaceRole) (*iamv1alpha2.WorkspaceRole, error)
	CreateOrUpdateGlobalRole(globalRole *iamv1alpha2.GlobalRole) (*iamv1alpha2.GlobalRole, error)
	DeleteWorkspaceRole(workspace string, name string) error
	DeleteGlobalRole(name string) error
	CreateOrUpdateClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error)
	DeleteClusterRole(name string) error
	GetClusterRole(name string) (*rbacv1.ClusterRole, error)
	GetNamespaceRole(namespace string, name string) (*rbacv1.Role, error)
	CreateOrUpdateNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error)
	DeleteNamespaceRole(namespace string, name string) error
	CreateOrUpdateWorkspaceRoleBinding(username string, workspace string, role string) error
	RemoveUserFromWorkspace(username string, workspace string) error
	CreateOrUpdateNamespaceRoleBinding(username string, namespace string, role string) error
	RemoveUserFromNamespace(username string, namespace string) error
	CreateOrUpdateClusterRoleBinding(username string, role string) error
	RemoveUserFromCluster(username string) error
	GetControlledNamespace(devops string) (string, error)
}

type amOperator struct {
	resourceGetter *resourcev1alpha3.ResourceGetter
	ksclient       kubesphere.Interface
	k8sclient      kubernetes.Interface
}

func NewReadOnlyOperator(factory informers.InformerFactory) AccessManagementInterface {
	return &amOperator{
		resourceGetter: resourcev1alpha3.NewResourceGetter(factory),
	}
}

func NewOperator(factory informers.InformerFactory, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) AccessManagementInterface {
	return &amOperator{
		resourceGetter: resourcev1alpha3.NewResourceGetter(factory),
		ksclient:       ksclient,
		k8sclient:      k8sclient,
	}
}

func (am *amOperator) GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error) {

	userRoleBindings, err := am.ListGlobalRoleBindings(username)

	if len(userRoleBindings) > 0 {
		role, err := am.GetGlobalRole(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(userRoleBindings) > 1 {
			klog.Warningf("conflict global role binding, username: %s", username)
		}

		out := role.DeepCopy()
		if out.Annotations == nil {
			out.Annotations = make(map[string]string, 0)
		}
		out.Annotations[iamv1alpha2.GlobalRoleAnnotation] = role.Name
		return out, nil
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

	userRoleBindings, err := am.ListWorkspaceRoleBindings(username, workspace)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		role, err := am.GetWorkspaceRole(workspace, userRoleBindings[0].RoleRef.Name)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if len(userRoleBindings) > 1 {
			klog.Warningf("conflict workspace role binding, username: %s", username)
		}

		out := role.DeepCopy()
		if out.Annotations == nil {
			out.Annotations = make(map[string]string, 0)
		}
		out.Annotations[iamv1alpha2.WorkspaceRoleAnnotation] = role.Name
		return out, nil
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
	userRoleBindings, err := am.ListRoleBindings(username, namespace)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		role, err := am.GetNamespaceRole(namespace, userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(userRoleBindings) > 1 {
			klog.Warningf("conflict role binding, username: %s", username)
		}

		out := role.DeepCopy()
		if out.Annotations == nil {
			out.Annotations = make(map[string]string, 0)
		}
		out.Annotations[iamv1alpha2.RoleAnnotation] = role.Name
		return out, nil
	}

	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: rbacv1.SchemeGroupVersion.Group,
			Kind:  iamv1alpha2.ResourceKindRoleBinding,
		},
		Message: fmt.Sprintf("role binding not found for %s in %s", username, namespace),
	}}

	return nil, err
}

func (am *amOperator) GetClusterRoleOfUser(username string) (*rbacv1.ClusterRole, error) {
	userRoleBindings, err := am.ListClusterRoleBindings(username)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		role, err := am.GetClusterRole(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if len(userRoleBindings) > 1 {
			klog.Warningf("conflict cluster role binding, username: %s", username)
		}

		out := role.DeepCopy()
		if out.Annotations == nil {
			out.Annotations = make(map[string]string, 0)
		}
		out.Annotations[iamv1alpha2.ClusterRoleAnnotation] = role.Name
		return out, nil
	}
	err = &errors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Group: rbacv1.SchemeGroupVersion.Group,
			Kind:  "ClusterRoleBinding",
		},
		Message: fmt.Sprintf("cluster role binding not found for %s", username),
	}}

	return nil, err
}

func (am *amOperator) ListWorkspaceRoleBindings(username, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error) {
	roleBindings, err := am.resourceGetter.List(iamv1alpha2.ResourcesPluralWorkspaceRoleBinding, "", query.New())

	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.WorkspaceRoleBinding, 0)

	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*iamv1alpha2.WorkspaceRoleBinding)
		inSpecifiedWorkspace := workspace == "" || roleBinding.Labels[tenantv1alpha1.WorkspaceLabel] == workspace
		if contains(roleBinding.Subjects, username) && inSpecifiedWorkspace {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error) {

	roleBindings, err := am.resourceGetter.List(iamv1alpha2.ResourcesPluralClusterRoleBinding, "", query.New())

	if err != nil {
		return nil, err
	}

	result := make([]*rbacv1.ClusterRoleBinding, 0)

	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.ClusterRoleBinding)
		if contains(roleBinding.Subjects, username) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error) {
	roleBindings, err := am.resourceGetter.List(iamv1alpha2.ResourcesPluralGlobalRoleBinding, "", query.New())

	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.GlobalRoleBinding, 0)

	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*iamv1alpha2.GlobalRoleBinding)
		if contains(roleBinding.Subjects, username) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListRoleBindings(username, namespace string) ([]*rbacv1.RoleBinding, error) {

	roleBindings, err := am.resourceGetter.List(iamv1alpha2.ResourcesPluralRoleBinding, namespace, query.New())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := make([]*rbacv1.RoleBinding, 0)

	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.RoleBinding)
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

func (am *amOperator) GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error) {
	obj, err := am.resourceGetter.Get(iamv1alpha2.ResourcesPluralGlobalRole, "", globalRole)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*iamv1alpha2.GlobalRole), nil
}

func (am *amOperator) CreateOrUpdateGlobalRoleBinding(username string, globalRole string) error {

	_, err := am.GetGlobalRole(globalRole)

	if err != nil {
		klog.Error(err)
		return err
	}

	roleBindings, err := am.ListGlobalRoleBindings(username)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if globalRole == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.ksclient.IamV1alpha2().GlobalRoleBindings().Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	globalRoleBinding := iamv1alpha2.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", username, globalRole),
			Labels: map[string]string{iamv1alpha2.UserReferenceLabel: username},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     username,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     globalRole,
		},
	}

	if _, err := am.ksclient.IamV1alpha2().GlobalRoleBindings().Create(&globalRoleBinding); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) CreateOrUpdateWorkspaceRole(workspace string, workspaceRole *iamv1alpha2.WorkspaceRole) (*iamv1alpha2.WorkspaceRole, error) {

	if workspaceRole.Labels == nil {
		workspaceRole.Labels = make(map[string]string, 0)
	}

	workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	workspaceRole.Rules = make([]rbacv1.PolicyRule, 0)

	var aggregateRoles []string
	if err := json.Unmarshal([]byte(workspaceRole.Annotations[iamv1alpha2.AggregationRolesAnnotation]), &aggregateRoles); err == nil {

		for _, roleName := range aggregateRoles {

			role, err := am.GetWorkspaceRole("", roleName)

			if err != nil {
				klog.Error(err)
				return nil, err
			}

			workspaceRole.Rules = append(workspaceRole.Rules, role.Rules...)
		}
	}

	old, err := am.GetWorkspaceRole("", workspaceRole.Name)

	if err != nil && !errors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	var created *iamv1alpha2.WorkspaceRole
	if old != nil {
		created, err = am.ksclient.IamV1alpha2().WorkspaceRoles().Update(workspaceRole)
	} else {
		created, err = am.ksclient.IamV1alpha2().WorkspaceRoles().Create(workspaceRole)
	}

	return created, err
}

func (am *amOperator) CreateOrUpdateWorkspaceRoleBinding(username string, workspace string, role string) error {

	_, err := am.GetWorkspaceRole(workspace, role)

	if err != nil {
		klog.Error(err)
		return err
	}

	roleBindings, err := am.ListWorkspaceRoleBindings(username, workspace)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	roleBinding := iamv1alpha2.WorkspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s", workspace, username, role),
			Labels: map[string]string{iamv1alpha2.UserReferenceLabel: username,
				tenantv1alpha1.WorkspaceLabel: workspace},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     username,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindWorkspaceRole,
			Name:     role,
		},
	}

	if _, err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Create(&roleBinding); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) CreateOrUpdateClusterRoleBinding(username string, role string) error {

	_, err := am.GetClusterRole(role)

	if err != nil {
		klog.Error(err)
		return err
	}

	roleBindings, err := am.ListClusterRoleBindings(username)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.k8sclient.RbacV1().ClusterRoleBindings().Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	roleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", username, role),
			Labels: map[string]string{iamv1alpha2.UserReferenceLabel: username},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     username,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindClusterRole,
			Name:     role,
		},
	}

	if _, err := am.k8sclient.RbacV1().ClusterRoleBindings().Create(&roleBinding); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) CreateOrUpdateNamespaceRoleBinding(username string, namespace string, role string) error {

	_, err := am.GetNamespaceRole(namespace, role)

	if err != nil {
		klog.Error(err)
		return err
	}

	roleBindings, err := am.ListRoleBindings(username, namespace)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.k8sclient.RbacV1().RoleBindings(namespace).Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", username, role),
			Labels: map[string]string{iamv1alpha2.UserReferenceLabel: username},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     username,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindRole,
			Name:     role,
		},
	}

	if _, err := am.k8sclient.RbacV1().RoleBindings(namespace).Create(&roleBinding); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) RemoveUserFromWorkspace(username string, workspace string) error {

	roleBindings, err := am.ListWorkspaceRoleBindings(username, workspace)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (am *amOperator) RemoveUserFromNamespace(username string, namespace string) error {

	roleBindings, err := am.ListRoleBindings(username, namespace)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.k8sclient.RbacV1().RoleBindings(namespace).Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (am *amOperator) RemoveUserFromCluster(username string) error {

	roleBindings, err := am.ListClusterRoleBindings(username)

	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.k8sclient.RbacV1().ClusterRoleBindings().Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (am *amOperator) CreateOrUpdateGlobalRole(globalRole *iamv1alpha2.GlobalRole) (*iamv1alpha2.GlobalRole, error) {

	globalRole.Rules = make([]rbacv1.PolicyRule, 0)

	var aggregateRoles []string
	if err := json.Unmarshal([]byte(globalRole.Annotations[iamv1alpha2.AggregationRolesAnnotation]), &aggregateRoles); err == nil {

		for _, roleName := range aggregateRoles {

			role, err := am.GetGlobalRole(roleName)

			if err != nil {
				klog.Error(err)
				return nil, err
			}

			globalRole.Rules = append(globalRole.Rules, role.Rules...)
		}
	}

	old, err := am.GetGlobalRole(globalRole.Name)

	if err != nil && !errors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	var created *iamv1alpha2.GlobalRole
	if old != nil {
		created, err = am.ksclient.IamV1alpha2().GlobalRoles().Update(globalRole)
	} else {
		created, err = am.ksclient.IamV1alpha2().GlobalRoles().Create(globalRole)
	}

	return created, err
}

func (am *amOperator) CreateOrUpdateClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {

	clusterRole.Rules = make([]rbacv1.PolicyRule, 0)

	var aggregateRoles []string
	if err := json.Unmarshal([]byte(clusterRole.Annotations[iamv1alpha2.AggregationRolesAnnotation]), &aggregateRoles); err == nil {

		for _, roleName := range aggregateRoles {

			role, err := am.GetClusterRole(roleName)

			if err != nil {
				klog.Error(err)
				return nil, err
			}

			clusterRole.Rules = append(clusterRole.Rules, role.Rules...)
		}
	}

	old, err := am.GetClusterRole(clusterRole.Name)

	if err != nil && !errors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	var created *rbacv1.ClusterRole
	if old != nil {
		created, err = am.k8sclient.RbacV1().ClusterRoles().Update(clusterRole)
	} else {
		created, err = am.k8sclient.RbacV1().ClusterRoles().Create(clusterRole)
	}

	return created, err
}

func (am *amOperator) CreateOrUpdateNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error) {

	role.Rules = make([]rbacv1.PolicyRule, 0)
	role.Namespace = namespace

	var aggregateRoles []string
	if err := json.Unmarshal([]byte(role.Annotations[iamv1alpha2.AggregationRolesAnnotation]), &aggregateRoles); err == nil {

		for _, roleName := range aggregateRoles {

			role, err := am.GetNamespaceRole(namespace, roleName)

			if err != nil {
				klog.Error(err)
				return nil, err
			}

			role.Rules = append(role.Rules, role.Rules...)
		}
	}

	old, err := am.GetNamespaceRole(namespace, role.Name)

	if err != nil && !errors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	var created *rbacv1.Role
	if old != nil {
		created, err = am.k8sclient.RbacV1().Roles(namespace).Update(role)
	} else {
		created, err = am.k8sclient.RbacV1().Roles(namespace).Create(role)
	}

	return created, err
}

func (am *amOperator) DeleteWorkspaceRole(workspace string, name string) error {
	workspaceRole, err := am.GetWorkspaceRole(workspace, name)
	if err != nil {
		return err
	}
	return am.ksclient.IamV1alpha2().WorkspaceRoles().Delete(workspaceRole.Name, metav1.NewDeleteOptions(0))
}

func (am *amOperator) DeleteGlobalRole(name string) error {
	return am.ksclient.IamV1alpha2().GlobalRoles().Delete(name, metav1.NewDeleteOptions(0))
}

func (am *amOperator) DeleteClusterRole(name string) error {
	return am.k8sclient.RbacV1().ClusterRoles().Delete(name, metav1.NewDeleteOptions(0))
}
func (am *amOperator) DeleteNamespaceRole(namespace string, name string) error {
	return am.k8sclient.RbacV1().Roles(namespace).Delete(name, metav1.NewDeleteOptions(0))
}

// GetRoleReferenceRules attempts to resolve the RoleBinding or ClusterRoleBinding.
func (am *amOperator) GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (string, []rbacv1.PolicyRule, error) {
	switch roleRef.Kind {
	case iamv1alpha2.ResourceKindRole:
		role, err := am.GetNamespaceRole(namespace, roleRef.Name)
		if err != nil {
			return "", nil, err
		}

		return role.Annotations[iamv1alpha2.RegoOverrideAnnotation], role.Rules, nil
	case iamv1alpha2.ResourceKindClusterRole:
		clusterRole, err := am.GetClusterRole(roleRef.Name)
		if err != nil {
			return "", nil, err
		}
		return clusterRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], clusterRole.Rules, nil
	case iamv1alpha2.ResourceKindGlobalRole:
		globalRole, err := am.GetGlobalRole(roleRef.Name)
		if err != nil {
			return "", nil, err
		}
		return globalRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], globalRole.Rules, nil
	case iamv1alpha2.ResourceKindWorkspaceRole:
		workspaceRole, err := am.GetWorkspaceRole("", roleRef.Name)
		if err != nil {
			return "", nil, err
		}
		return workspaceRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], workspaceRole.Rules, nil
	default:
		return "", nil, fmt.Errorf("unsupported role reference kind: %q", roleRef.Kind)
	}
}

func (am *amOperator) GetWorkspaceRole(workspace string, name string) (*iamv1alpha2.WorkspaceRole, error) {
	obj, err := am.resourceGetter.Get(iamv1alpha2.ResourcesPluralWorkspaceRole, "", name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaceRole := obj.(*iamv1alpha2.WorkspaceRole)

	if workspace != "" && workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularWorkspaceRole), name)
		klog.Error(err)
		return nil, err
	}

	return workspaceRole, nil
}

func (am *amOperator) GetNamespaceRole(namespace string, name string) (*rbacv1.Role, error) {
	obj, err := am.resourceGetter.Get(iamv1alpha2.ResourcesPluralRole, namespace, name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.Role), nil
}

func (am *amOperator) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	obj, err := am.resourceGetter.Get(iamv1alpha2.ResourcesPluralClusterRole, "", name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.ClusterRole), nil
}
func (am *amOperator) GetControlledNamespace(devops string) (string, error) {
	obj, err := am.resourceGetter.Get(devopsv1alpha3.ResourcePluralDevOpsProject, "", devops)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	devopsProject := obj.(*devopsv1alpha3.DevOpsProject)

	return devopsProject.Status.AdminNamespace, nil
}
