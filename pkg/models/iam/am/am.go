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
	"context"
	"encoding/json"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/role"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/rolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerolebinding"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type AccessManagementInterface interface {
	GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error)
	GetWorkspaceRoleOfUser(username string, groups []string, workspace string) ([]*iamv1alpha2.WorkspaceRole, error)
	GetClusterRoleOfUser(username string) (*rbacv1.ClusterRole, error)
	GetNamespaceRoleOfUser(username string, groups []string, namespace string) ([]*rbacv1.Role, error)
	ListRoles(namespace string, query *query.Query) (*api.ListResult, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	ListWorkspaceRoles(query *query.Query) (*api.ListResult, error)
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)
	ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error)
	ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error)
	ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error)
	ListRoleBindings(username string, groups []string, namespace string) ([]*rbacv1.RoleBinding, error)
	GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (string, []rbacv1.PolicyRule, error)
	GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error)
	GetWorkspaceRole(workspace string, name string) (*iamv1alpha2.WorkspaceRole, error)
	CreateGlobalRoleBinding(username string, globalRole string) error
	CreateOrUpdateWorkspaceRole(workspace string, workspaceRole *iamv1alpha2.WorkspaceRole) (*iamv1alpha2.WorkspaceRole, error)
	PatchWorkspaceRole(workspace string, workspaceRole *iamv1alpha2.WorkspaceRole) (*iamv1alpha2.WorkspaceRole, error)
	CreateOrUpdateGlobalRole(globalRole *iamv1alpha2.GlobalRole) (*iamv1alpha2.GlobalRole, error)
	PatchGlobalRole(globalRole *iamv1alpha2.GlobalRole) (*iamv1alpha2.GlobalRole, error)
	DeleteWorkspaceRole(workspace string, name string) error
	DeleteGlobalRole(name string) error
	CreateOrUpdateClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error)
	DeleteClusterRole(name string) error
	GetClusterRole(name string) (*rbacv1.ClusterRole, error)
	GetNamespaceRole(namespace string, name string) (*rbacv1.Role, error)
	CreateOrUpdateNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error)
	DeleteNamespaceRole(namespace string, name string) error
	CreateUserWorkspaceRoleBinding(username string, workspace string, role string) error
	RemoveUserFromWorkspace(username string, workspace string) error
	CreateNamespaceRoleBinding(username string, namespace string, role string) error
	RemoveUserFromNamespace(username string, namespace string) error
	CreateClusterRoleBinding(username string, role string) error
	RemoveUserFromCluster(username string) error
	GetDevOpsRelatedNamespace(devops string) (string, error)
	GetNamespaceControlledWorkspace(namespace string) (string, error)
	GetDevOpsControlledWorkspace(devops string) (string, error)
	PatchNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error)
	PatchClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error)
	ListGroupRoleBindings(workspace string, query *query.Query) ([]*rbacv1.RoleBinding, error)
	CreateRoleBinding(namespace string, roleBinding *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error)
	DeleteRoleBinding(namespace, name string) error
	ListGroupWorkspaceRoleBindings(workspace string, query *query.Query) (*api.ListResult, error)
	CreateWorkspaceRoleBinding(workspace string, roleBinding *iamv1alpha2.WorkspaceRoleBinding) (*iamv1alpha2.WorkspaceRoleBinding, error)
	DeleteWorkspaceRoleBinding(workspaceName, name string) error
}

type amOperator struct {
	globalRoleBindingGetter    resourcev1alpha3.Interface
	workspaceRoleBindingGetter resourcev1alpha3.Interface
	clusterRoleBindingGetter   resourcev1alpha3.Interface
	roleBindingGetter          resourcev1alpha3.Interface
	globalRoleGetter           resourcev1alpha3.Interface
	workspaceRoleGetter        resourcev1alpha3.Interface
	clusterRoleGetter          resourcev1alpha3.Interface
	roleGetter                 resourcev1alpha3.Interface
	devopsProjectLister        devopslisters.DevOpsProjectLister
	namespaceLister            listersv1.NamespaceLister
	ksclient                   kubesphere.Interface
	k8sclient                  kubernetes.Interface
}

func NewReadOnlyOperator(factory informers.InformerFactory) AccessManagementInterface {
	return &amOperator{
		globalRoleBindingGetter:    globalrolebinding.New(factory.KubeSphereSharedInformerFactory()),
		workspaceRoleBindingGetter: workspacerolebinding.New(factory.KubeSphereSharedInformerFactory()),
		clusterRoleBindingGetter:   clusterrolebinding.New(factory.KubernetesSharedInformerFactory()),
		roleBindingGetter:          rolebinding.New(factory.KubernetesSharedInformerFactory()),
		globalRoleGetter:           globalrole.New(factory.KubeSphereSharedInformerFactory()),
		workspaceRoleGetter:        workspacerole.New(factory.KubeSphereSharedInformerFactory()),
		clusterRoleGetter:          clusterrole.New(factory.KubernetesSharedInformerFactory()),
		roleGetter:                 role.New(factory.KubernetesSharedInformerFactory()),
		devopsProjectLister:        factory.KubeSphereSharedInformerFactory().Devops().V1alpha3().DevOpsProjects().Lister(),
		namespaceLister:            factory.KubernetesSharedInformerFactory().Core().V1().Namespaces().Lister(),
	}
}

func NewOperator(ksClient kubesphere.Interface, k8sClient kubernetes.Interface, factory informers.InformerFactory) AccessManagementInterface {
	amOperator := NewReadOnlyOperator(factory).(*amOperator)
	amOperator.ksclient = ksClient
	amOperator.k8sclient = k8sClient
	return amOperator
}

func (am *amOperator) GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error) {
	globalRoleBindings, err := am.ListGlobalRoleBindings(username)
	if len(globalRoleBindings) > 0 {
		// Usually, only one globalRoleBinding will be found which is created from ks-console.
		if len(globalRoleBindings) > 1 {
			klog.Warningf("conflict global role binding, username: %s", username)
		}
		globalRole, err := am.GetGlobalRole(globalRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return globalRole, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularGlobalRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

func (am *amOperator) GetWorkspaceRoleOfUser(username string, groups []string, workspace string) ([]*iamv1alpha2.WorkspaceRole, error) {

	userRoleBindings, err := am.ListWorkspaceRoleBindings(username, groups, workspace)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		roles := make([]*iamv1alpha2.WorkspaceRole, len(userRoleBindings))
		for i, roleBinding := range userRoleBindings {
			role, err := am.GetWorkspaceRole(workspace, roleBinding.RoleRef.Name)

			if err != nil {
				klog.Error(err)
				return nil, err
			}

			out := role.DeepCopy()
			if out.Annotations == nil {
				out.Annotations = make(map[string]string, 0)
			}
			out.Annotations[iamv1alpha2.WorkspaceRoleAnnotation] = role.Name

			roles[i] = out
		}

		if len(userRoleBindings) > 1 {
			klog.Infof("conflict workspace role binding, username: %s", username)
		}

		return roles, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularWorkspaceRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

func (am *amOperator) GetNamespaceRoleOfUser(username string, groups []string, namespace string) ([]*rbacv1.Role, error) {

	userRoleBindings, err := am.ListRoleBindings(username, groups, namespace)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		roles := make([]*rbacv1.Role, len(userRoleBindings))
		for i, roleBinding := range userRoleBindings {
			role, err := am.GetNamespaceRole(namespace, roleBinding.RoleRef.Name)
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			out := role.DeepCopy()
			if out.Annotations == nil {
				out.Annotations = make(map[string]string, 0)
			}
			out.Annotations[iamv1alpha2.RoleAnnotation] = role.Name

			roles[i] = out
		}

		if len(userRoleBindings) > 1 {
			klog.Infof("conflict role binding, username: %s", username)
		}
		return roles, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularRoleBinding), username)
	klog.V(4).Info(err)
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

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularClusterRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

func (am *amOperator) ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error) {
	roleBindings, err := am.workspaceRoleBindingGetter.List("", query.New())

	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.WorkspaceRoleBinding, 0)

	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*iamv1alpha2.WorkspaceRoleBinding)
		inSpecifiedWorkspace := workspace == "" || roleBinding.Labels[tenantv1alpha1.WorkspaceLabel] == workspace
		if contains(roleBinding.Subjects, username, groups) && inSpecifiedWorkspace {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error) {

	roleBindings, err := am.clusterRoleBindingGetter.List("", query.New())
	if err != nil {
		return nil, err
	}

	result := make([]*rbacv1.ClusterRoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.ClusterRoleBinding)
		if contains(roleBinding.Subjects, username, nil) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error) {
	roleBindings, err := am.globalRoleBindingGetter.List("", query.New())
	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.GlobalRoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*iamv1alpha2.GlobalRoleBinding)
		if contains(roleBinding.Subjects, username, nil) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListRoleBindings(username string, groups []string, namespace string) ([]*rbacv1.RoleBinding, error) {
	roleBindings, err := am.roleBindingGetter.List(namespace, query.New())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := make([]*rbacv1.RoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.RoleBinding)
		if contains(roleBinding.Subjects, username, groups) {
			result = append(result, roleBinding)
		}
	}
	return result, nil
}

func contains(subjects []rbacv1.Subject, username string, groups []string) bool {
	// if username is nil means list all role bindings
	if username == "" {
		return true
	}
	for _, subject := range subjects {
		if subject.Kind == rbacv1.UserKind && subject.Name == username {
			return true
		}
		if subject.Kind == rbacv1.GroupKind && sliceutil.HasString(groups, subject.Name) {
			return true
		}
	}
	return false
}

func (am *amOperator) ListRoles(namespace string, query *query.Query) (*api.ListResult, error) {
	return am.roleGetter.List(namespace, query)
}

func (am *amOperator) ListClusterRoles(query *query.Query) (*api.ListResult, error) {
	return am.clusterRoleGetter.List("", query)
}

func (am *amOperator) ListWorkspaceRoles(queryParam *query.Query) (*api.ListResult, error) {
	return am.workspaceRoleGetter.List("", queryParam)
}

func (am *amOperator) ListGlobalRoles(query *query.Query) (*api.ListResult, error) {
	return am.globalRoleGetter.List("", query)
}

func (am *amOperator) GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error) {
	obj, err := am.globalRoleGetter.Get("", globalRole)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*iamv1alpha2.GlobalRole), nil
}

func (am *amOperator) CreateGlobalRoleBinding(username string, role string) error {
	_, err := am.GetGlobalRole(role)
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
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.ksclient.IamV1alpha2().GlobalRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     role,
		},
	}

	if _, err := am.ksclient.IamV1alpha2().GlobalRoleBindings().Create(context.Background(), &globalRoleBinding, metav1.CreateOptions{}); err != nil {
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
	if aggregateRoles := am.getAggregateRoles(workspaceRole.ObjectMeta); aggregateRoles != nil {
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetWorkspaceRole("", roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			workspaceRole.Rules = append(workspaceRole.Rules, aggregationRole.Rules...)
		}
	}
	var created *iamv1alpha2.WorkspaceRole
	var err error
	if workspaceRole.ResourceVersion != "" {
		created, err = am.ksclient.IamV1alpha2().WorkspaceRoles().Update(context.Background(), workspaceRole, metav1.UpdateOptions{})
	} else {
		created, err = am.ksclient.IamV1alpha2().WorkspaceRoles().Create(context.Background(), workspaceRole, metav1.CreateOptions{})
	}

	return created, err
}

func (am *amOperator) PatchGlobalRole(globalRole *iamv1alpha2.GlobalRole) (*iamv1alpha2.GlobalRole, error) {
	old, err := am.GetGlobalRole(globalRole.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// rules cannot be override
	globalRole.Rules = old.Rules
	// aggregate roles if annotation has change
	if aggregateRoles := am.getAggregateRoles(globalRole.ObjectMeta); aggregateRoles != nil {
		globalRole.Rules = make([]rbacv1.PolicyRule, 0)
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetGlobalRole(roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			globalRole.Rules = append(globalRole.Rules, aggregationRole.Rules...)
		}
	}

	data, err := json.Marshal(globalRole)
	if err != nil {
		return nil, err
	}

	return am.ksclient.IamV1alpha2().GlobalRoles().Patch(context.Background(), globalRole.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (am *amOperator) getAggregateRoles(obj metav1.ObjectMeta) []string {
	if aggregateRolesAnnotation := obj.Annotations[iamv1alpha2.AggregationRolesAnnotation]; aggregateRolesAnnotation != "" {
		var aggregateRoles []string
		if err := json.Unmarshal([]byte(aggregateRolesAnnotation), &aggregateRoles); err != nil {
			klog.Warningf("invalid aggregation role annotation found %+v", obj)
		}
		return aggregateRoles
	}
	return nil
}

func (am *amOperator) PatchWorkspaceRole(workspace string, workspaceRole *iamv1alpha2.WorkspaceRole) (*iamv1alpha2.WorkspaceRole, error) {
	old, err := am.GetWorkspaceRole(workspace, workspaceRole.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// workspace label cannot be override
	if workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] != "" {
		workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	}

	// rules cannot be override
	workspaceRole.Rules = old.Rules
	// aggregate roles if annotation has change
	if aggregateRoles := am.getAggregateRoles(workspaceRole.ObjectMeta); aggregateRoles != nil {
		workspaceRole.Rules = make([]rbacv1.PolicyRule, 0)
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetWorkspaceRole("", roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			workspaceRole.Rules = append(workspaceRole.Rules, aggregationRole.Rules...)
		}
	}

	data, err := json.Marshal(workspaceRole)
	if err != nil {
		return nil, err
	}

	return am.ksclient.IamV1alpha2().WorkspaceRoles().Patch(context.Background(), workspaceRole.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (am *amOperator) PatchNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error) {
	old, err := am.GetNamespaceRole(namespace, role.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// rules cannot be override
	role.Rules = old.Rules
	// aggregate roles if annotation has change
	if aggregateRoles := am.getAggregateRoles(role.ObjectMeta); aggregateRoles != nil {
		role.Rules = make([]rbacv1.PolicyRule, 0)
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetNamespaceRole(namespace, roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			role.Rules = append(role.Rules, aggregationRole.Rules...)
		}
	}

	data, err := json.Marshal(role)
	if err != nil {
		return nil, err
	}

	return am.k8sclient.RbacV1().Roles(namespace).Patch(context.Background(), role.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (am *amOperator) PatchClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	old, err := am.GetClusterRole(clusterRole.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// rules cannot be override
	clusterRole.Rules = old.Rules
	// aggregate roles if annotation has change
	if aggregateRoles := am.getAggregateRoles(clusterRole.ObjectMeta); aggregateRoles != nil {
		clusterRole.Rules = make([]rbacv1.PolicyRule, 0)
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetClusterRole(roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			clusterRole.Rules = append(clusterRole.Rules, aggregationRole.Rules...)
		}
	}

	data, err := json.Marshal(clusterRole)
	if err != nil {
		return nil, err
	}

	return am.k8sclient.RbacV1().ClusterRoles().Patch(context.Background(), clusterRole.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (am *amOperator) CreateUserWorkspaceRoleBinding(username string, workspace string, role string) error {
	_, err := am.GetWorkspaceRole(workspace, role)
	if err != nil {
		klog.Error(err)
		return err
	}

	roleBindings, err := am.ListWorkspaceRoleBindings(username, nil, workspace)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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
			Name: fmt.Sprintf("%s-%s", username, role),
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

	if _, err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Create(context.Background(), &roleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) CreateClusterRoleBinding(username string, role string) error {
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
		err := am.k8sclient.RbacV1().ClusterRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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

	if _, err := am.k8sclient.RbacV1().ClusterRoleBindings().Create(context.Background(), &roleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) CreateNamespaceRoleBinding(username string, namespace string, role string) error {

	_, err := am.GetNamespaceRole(namespace, role)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Don't pass user's groups.
	roleBindings, err := am.ListRoleBindings(username, nil, namespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		if role == roleBinding.RoleRef.Name {
			return nil
		}
		err := am.k8sclient.RbacV1().RoleBindings(namespace).Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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

	if _, err := am.k8sclient.RbacV1().RoleBindings(namespace).Create(context.Background(), &roleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (am *amOperator) RemoveUserFromWorkspace(username string, workspace string) error {

	roleBindings, err := am.ListWorkspaceRoleBindings(username, nil, workspace)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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

	roleBindings, err := am.ListRoleBindings(username, nil, namespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.k8sclient.RbacV1().RoleBindings(namespace).Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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
		err := am.k8sclient.RbacV1().ClusterRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
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
	if aggregateRoles := am.getAggregateRoles(globalRole.ObjectMeta); aggregateRoles != nil {
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetGlobalRole(roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			globalRole.Rules = append(globalRole.Rules, aggregationRole.Rules...)
		}
	}
	var created *iamv1alpha2.GlobalRole
	var err error
	if globalRole.ResourceVersion != "" {
		created, err = am.ksclient.IamV1alpha2().GlobalRoles().Update(context.Background(), globalRole, metav1.UpdateOptions{})
	} else {
		created, err = am.ksclient.IamV1alpha2().GlobalRoles().Create(context.Background(), globalRole, metav1.CreateOptions{})
	}
	return created, err
}

func (am *amOperator) CreateOrUpdateClusterRole(clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	clusterRole.Rules = make([]rbacv1.PolicyRule, 0)
	if aggregateRoles := am.getAggregateRoles(clusterRole.ObjectMeta); aggregateRoles != nil {
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetClusterRole(roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			clusterRole.Rules = append(clusterRole.Rules, aggregationRole.Rules...)
		}
	}
	var created *rbacv1.ClusterRole
	var err error
	if clusterRole.ResourceVersion != "" {
		created, err = am.k8sclient.RbacV1().ClusterRoles().Update(context.Background(), clusterRole, metav1.UpdateOptions{})
	} else {
		created, err = am.k8sclient.RbacV1().ClusterRoles().Create(context.Background(), clusterRole, metav1.CreateOptions{})
	}
	return created, err
}

func (am *amOperator) CreateOrUpdateNamespaceRole(namespace string, role *rbacv1.Role) (*rbacv1.Role, error) {
	role.Rules = make([]rbacv1.PolicyRule, 0)
	role.Namespace = namespace
	if aggregateRoles := am.getAggregateRoles(role.ObjectMeta); aggregateRoles != nil {
		for _, roleName := range aggregateRoles {
			aggregationRole, err := am.GetNamespaceRole(namespace, roleName)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			role.Rules = append(role.Rules, aggregationRole.Rules...)
		}
	}
	var created *rbacv1.Role
	var err error
	if role.ResourceVersion != "" {
		created, err = am.k8sclient.RbacV1().Roles(namespace).Update(context.Background(), role, metav1.UpdateOptions{})
	} else {
		created, err = am.k8sclient.RbacV1().Roles(namespace).Create(context.Background(), role, metav1.CreateOptions{})
	}

	return created, err
}

func (am *amOperator) DeleteWorkspaceRole(workspace string, name string) error {
	workspaceRole, err := am.GetWorkspaceRole(workspace, name)
	if err != nil {
		return err
	}
	return am.ksclient.IamV1alpha2().WorkspaceRoles().Delete(context.Background(), workspaceRole.Name, *metav1.NewDeleteOptions(0))
}

func (am *amOperator) DeleteGlobalRole(name string) error {
	return am.ksclient.IamV1alpha2().GlobalRoles().Delete(context.Background(), name, *metav1.NewDeleteOptions(0))
}

func (am *amOperator) DeleteClusterRole(name string) error {
	return am.k8sclient.RbacV1().ClusterRoles().Delete(context.Background(), name, *metav1.NewDeleteOptions(0))
}
func (am *amOperator) DeleteNamespaceRole(namespace string, name string) error {
	return am.k8sclient.RbacV1().Roles(namespace).Delete(context.Background(), name, *metav1.NewDeleteOptions(0))
}

// GetRoleReferenceRules attempts to resolve the RoleBinding or ClusterRoleBinding.
func (am *amOperator) GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (regoPolicy string, rules []rbacv1.PolicyRule, err error) {

	empty := make([]rbacv1.PolicyRule, 0)

	switch roleRef.Kind {
	case iamv1alpha2.ResourceKindRole:
		role, err := am.GetNamespaceRole(namespace, roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return role.Annotations[iamv1alpha2.RegoOverrideAnnotation], role.Rules, nil
	case iamv1alpha2.ResourceKindClusterRole:
		clusterRole, err := am.GetClusterRole(roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return clusterRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], clusterRole.Rules, nil
	case iamv1alpha2.ResourceKindGlobalRole:
		globalRole, err := am.GetGlobalRole(roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return globalRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], globalRole.Rules, nil
	case iamv1alpha2.ResourceKindWorkspaceRole:
		workspaceRole, err := am.GetWorkspaceRole("", roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return workspaceRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], workspaceRole.Rules, nil
	default:
		return "", nil, fmt.Errorf("unsupported role reference kind: %q", roleRef.Kind)
	}
}

func (am *amOperator) GetWorkspaceRole(workspace string, name string) (*iamv1alpha2.WorkspaceRole, error) {
	obj, err := am.workspaceRoleGetter.Get("", name)
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
	obj, err := am.roleGetter.Get(namespace, name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.Role), nil
}

func (am *amOperator) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	obj, err := am.clusterRoleGetter.Get("", name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.ClusterRole), nil
}
func (am *amOperator) GetDevOpsRelatedNamespace(devops string) (string, error) {
	devopsProject, err := am.devopsProjectLister.Get(devops)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return devopsProject.Status.AdminNamespace, nil
}

func (am *amOperator) GetDevOpsControlledWorkspace(devops string) (string, error) {
	devopsProject, err := am.devopsProjectLister.Get(devops)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return devopsProject.Labels[tenantv1alpha1.WorkspaceLabel], nil
}

func (am *amOperator) GetNamespaceControlledWorkspace(namespace string) (string, error) {
	ns, err := am.namespaceLister.Get(namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}
		klog.Error(err)
		return "", err
	}
	return ns.Labels[tenantv1alpha1.WorkspaceLabel], nil
}

func (am *amOperator) ListGroupWorkspaceRoleBindings(workspace string, query *query.Query) (*api.ListResult, error) {

	lableSelector, err := labels.ConvertSelectorToLabelsMap(query.LabelSelector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// workspace resources must be filtered by workspace
	wsSelector := labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}
	query.LabelSelector = labels.Merge(lableSelector, wsSelector).String()
	return am.workspaceRoleBindingGetter.List("", query)
}

func (am *amOperator) CreateWorkspaceRoleBinding(workspace string, roleBinding *iamv1alpha2.WorkspaceRoleBinding) (*iamv1alpha2.WorkspaceRoleBinding, error) {

	_, err := am.GetWorkspaceRole(workspace, roleBinding.RoleRef.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(roleBinding.Subjects) == 0 {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesPluralUser), "")
		return nil, err
	}

	roleBinding.GenerateName = fmt.Sprintf("%s-%s-", roleBinding.Subjects[0].Name, roleBinding.RoleRef.Name)

	if roleBinding.Labels == nil {
		roleBinding.Labels = map[string]string{}
	}

	if roleBinding.Subjects[0].Kind == rbacv1.GroupKind {
		roleBinding.Labels[iamv1alpha2.GroupReferenceLabel] = roleBinding.Subjects[0].Name
	} else if roleBinding.Subjects[0].Kind == rbacv1.UserKind {
		roleBinding.Labels[iamv1alpha2.UserReferenceLabel] = roleBinding.Subjects[0].Name
	}

	roleBinding.Labels[tenantv1alpha1.WorkspaceLabel] = workspace

	return am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Create(context.Background(), roleBinding, metav1.CreateOptions{})

}
func (am *amOperator) DeleteWorkspaceRoleBinding(workspaceName, name string) error {
	return am.ksclient.IamV1alpha2().WorkspaceRoleBindings().Delete(context.Background(), name, *metav1.NewDeleteOptions(0))
}

func (am *amOperator) ListGroupRoleBindings(workspace string, query *query.Query) ([]*rbacv1.RoleBinding, error) {
	namespaces, err := am.namespaceLister.List(labels.SelectorFromSet(labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}))
	if err != nil {
		return nil, err
	}
	result := make([]*rbacv1.RoleBinding, 0)
	for _, namespace := range namespaces {
		roleBindings, err := am.roleBindingGetter.List(namespace.Name, query)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		for _, obj := range roleBindings.Items {
			roleBinding := obj.(*rbacv1.RoleBinding)
			result = append(result, roleBinding)
		}
	}
	devOpsProjects, err := am.devopsProjectLister.List(labels.SelectorFromSet(labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}))
	if err != nil {
		return nil, err
	}
	for _, devOpsProject := range devOpsProjects {
		roleBindings, err := am.roleBindingGetter.List(devOpsProject.Name, query)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		for _, obj := range roleBindings.Items {
			roleBinding := obj.(*rbacv1.RoleBinding)
			result = append(result, roleBinding)
		}
	}
	return result, nil
}

func (am *amOperator) CreateRoleBinding(namespace string, roleBinding *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {

	_, err := am.GetNamespaceRole(namespace, roleBinding.RoleRef.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(roleBinding.Subjects) == 0 {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesPluralUser), "")
		return nil, err
	}

	roleBinding.GenerateName = fmt.Sprintf("%s-%s-", roleBinding.Subjects[0].Name, roleBinding.RoleRef.Name)

	if roleBinding.Labels == nil {
		roleBinding.Labels = map[string]string{}
	}

	if roleBinding.Subjects[0].Kind == rbacv1.GroupKind {
		roleBinding.Labels[iamv1alpha2.GroupReferenceLabel] = roleBinding.Subjects[0].Name
	} else if roleBinding.Subjects[0].Kind == rbacv1.UserKind {
		roleBinding.Labels[iamv1alpha2.UserReferenceLabel] = roleBinding.Subjects[0].Name
	}

	return am.k8sclient.RbacV1().RoleBindings(namespace).Create(context.Background(), roleBinding, metav1.CreateOptions{})
}

func (am *amOperator) DeleteRoleBinding(namespace, name string) error {
	return am.k8sclient.RbacV1().RoleBindings(namespace).Delete(context.Background(), name, *metav1.NewDeleteOptions(0))
}
