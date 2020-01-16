/*

 Copyright 2019 The KubeSphere Authors.

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
package iam

import (
	"fmt"
	"github.com/go-ldap/ldap"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/role"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"
)

const (
	ClusterRoleKind             = "ClusterRole"
	NamespaceAdminRoleBindName  = "admin"
	NamespaceViewerRoleBindName = "viewer"
)

type AccessManagementInterface interface {
	GetDevopsRoleSimpleRules(role string) []SimpleRule
	ListRoleBindings(namespace string, role string) ([]*rbacv1.RoleBinding, error)
	CreateClusterRoleBinding(username string, clusterRole string) error
}

type amOperator struct {
	informers informers.SharedInformerFactory
	resources resource.ResourceGetter
}

func newAMOperator(informers informers.SharedInformerFactory) *amOperator {
	resourceGetter := resource.ResourceGetter{}
	resourceGetter.Add(v1alpha2.Role, role.NewRoleSearcher(informers))
	resourceGetter.Add(v1alpha2.ClusterRoles, clusterrole.NewClusterRoleSearcher(informers))
	return &amOperator{informers: informers, resources: resourceGetter}
}

func (am *amOperator) GetDevopsRoleSimpleRules(role string) []SimpleRule {
	var rules []SimpleRule

	switch role {
	case "developer":
		rules = []SimpleRule{
			{Name: "pipelines", Actions: []string{"view", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "owner":
		rules = []SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"edit", "view", "delete"}},
		}
		break
	case "maintainer":
		rules = []SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "reporter":
		fallthrough
	default:
		rules = []SimpleRule{
			{Name: "pipelines", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	}
	return rules
}

// Get user roles in namespace
func (am *amOperator) GetUserRoles(namespace, username string) ([]*rbacv1.Role, error) {
	clusterRoleLister := am.informers.Rbac().V1().ClusterRoles().Lister()
	roleBindingLister := am.informers.Rbac().V1().RoleBindings().Lister()
	roleLister := am.informers.Rbac().V1().Roles().Lister()
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	roles := make([]*rbacv1.Role, 0)

	for _, roleBinding := range roleBindings {
		if k8sutil.ContainsUser(roleBinding.Subjects, username) {
			if roleBinding.RoleRef.Kind == ClusterRoleKind {
				clusterRole, err := clusterRoleLister.Get(roleBinding.RoleRef.Name)
				if err != nil {
					if apierrors.IsNotFound(err) {
						klog.Warningf("cluster role %s not found but bind user %s in namespace %s", roleBinding.RoleRef.Name, username, namespace)
						continue
					} else {
						klog.Errorln(err)
						return nil, err
					}
				}
				role := rbacv1.Role{}
				role.TypeMeta = clusterRole.TypeMeta
				role.ObjectMeta = clusterRole.ObjectMeta
				role.Rules = clusterRole.Rules
				role.Namespace = roleBinding.Namespace
				roles = append(roles, &role)
			} else {
				role, err := roleLister.Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name)

				if err != nil {
					if apierrors.IsNotFound(err) {
						klog.Warningf("namespace %s role %s not found, but bind user %s", namespace, roleBinding.RoleRef.Name, username)
						continue
					} else {
						klog.Errorln(err)
						return nil, err
					}
				}
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func (am *amOperator) GetUserClusterRoles(username string) (*rbacv1.ClusterRole, []*rbacv1.ClusterRole, error) {
	clusterRoleLister := am.informers.Rbac().V1().ClusterRoles().Lister()
	clusterRoleBindingLister := am.informers.Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, nil, err
	}

	clusterRoles := make([]*rbacv1.ClusterRole, 0)
	userFacingClusterRole := &rbacv1.ClusterRole{}
	for _, clusterRoleBinding := range clusterRoleBindings {
		if k8sutil.ContainsUser(clusterRoleBinding.Subjects, username) {
			clusterRole, err := clusterRoleLister.Get(clusterRoleBinding.RoleRef.Name)
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Warningf("cluster role %s not found but bind user %s", clusterRoleBinding.RoleRef.Name, username)
					continue
				} else {
					klog.Errorln(err)
					return nil, nil, err
				}
			}
			if clusterRoleBinding.Name == username {
				userFacingClusterRole = clusterRole
			}
			clusterRoles = append(clusterRoles, clusterRole)
		}
	}

	return userFacingClusterRole, clusterRoles, nil
}

func (am *amOperator) GetUserClusterRole(username string) (*rbacv1.ClusterRole, error) {
	userFacingClusterRole, _, err := am.GetUserClusterRoles(username)
	if err != nil {
		return nil, err
	}
	return userFacingClusterRole, nil
}

func (am *amOperator) GetUserClusterRules(username string) ([]rbacv1.PolicyRule, error) {
	_, clusterRoles, err := am.GetUserClusterRoles(username)

	if err != nil {
		return nil, err
	}

	rules := make([]rbacv1.PolicyRule, 0)
	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	return rules, nil
}

func (am *amOperator) GetUserRules(namespace, username string) ([]rbacv1.PolicyRule, error) {
	roles, err := am.GetUserRoles(namespace, username)

	if err != nil {
		return nil, err
	}

	rules := make([]rbacv1.PolicyRule, 0)
	for _, role := range roles {
		rules = append(rules, role.Rules...)
	}

	return rules, nil
}

func (am *amOperator) GetWorkspaceRoleBindings(workspace string) ([]*rbacv1.ClusterRoleBinding, error) {

	clusterRoleBindings, err := am.informers.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Errorln("get cluster role bindings", err)
		return nil, err
	}

	result := make([]*rbacv1.ClusterRoleBinding, 0)

	for _, roleBinding := range clusterRoleBindings {
		if k8sutil.IsControlledBy(roleBinding.OwnerReferences, "Workspace", workspace) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) GetWorkspaceRole(workspace, role string) (*rbacv1.ClusterRole, error) {
	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return nil, apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}
	role = fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))
	return am.informers.Rbac().V1().ClusterRoles().Lister().Get(role)
}

func (am *amOperator) GetUserWorkspaceRoleMap(username string) (map[string]string, error) {

	clusterRoleBindings, err := am.informers.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Errorln("get cluster role bindings", err)
		return nil, err
	}

	result := make(map[string]string, 0)

	for _, roleBinding := range clusterRoleBindings {
		if workspace := k8sutil.GetControlledWorkspace(roleBinding.OwnerReferences); workspace != "" &&
			k8sutil.ContainsUser(roleBinding.Subjects, username) {
			result[workspace] = roleBinding.RoleRef.Name
		}
	}

	return result, nil
}

func (am *amOperator) GetUserWorkspaceRole(workspace, username string) (*rbacv1.ClusterRole, error) {
	workspaceRoleMap, err := am.GetUserWorkspaceRoleMap(username)

	if err != nil {
		return nil, err
	}

	if workspaceRole := workspaceRoleMap[workspace]; workspaceRole != "" {
		return am.informers.Rbac().V1().ClusterRoles().Lister().Get(workspaceRole)
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{Resource: "workspace user"}, username)
}

func (am *amOperator) GetRoleBindings(namespace string, roleName string) ([]*rbacv1.RoleBinding, error) {
	roleBindingLister := am.informers.Rbac().V1().RoleBindings().Lister()
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	items := make([]*rbacv1.RoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if roleName == "" {
			items = append(items, roleBinding)
		} else if roleBinding.RoleRef.Name == roleName {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func (am *amOperator) GetClusterRoleBindings(clusterRoleName string) ([]*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingLister := am.informers.Rbac().V1().ClusterRoleBindings().Lister()
	roleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	items := make([]*rbacv1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindings {
		if roleBinding.RoleRef.Name == clusterRoleName {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func (am *amOperator) ListClusterRoleUsers(clusterRoleName string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	roleBindings, err := am.GetClusterRoleBindings(clusterRoleName)

	if err != nil {
		return nil, err
	}
	users := make([]*User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				user, err := GetUserInfo(subject.Name)
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					klog.Errorln(err)
					return nil, err
				}
				users = append(users, user)
			}
		}
	}

	// order & reverse
	sort.Slice(users, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		switch orderBy {
		default:
			fallthrough
		case v1alpha2.Name:
			return strings.Compare(users[i].Username, users[j].Username) <= 0
		}
	})

	result := make([]interface{}, 0)

	for i, d := range users {
		if i >= offset && (limit == -1 || len(result) < limit) {
			result = append(result, d)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(users)}, nil

}

func (am *amOperator) ListRoles(namespace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	return am.resources.ListResources(namespace, v1alpha2.Roles, conditions, orderBy, reverse, limit, offset)
}

func (am *amOperator) ListRoleBindings(namespace string, role string) ([]*rbacv1.RoleBinding, error) {
	rbs, err := am.informers.Rbac().V1().RoleBindings().Lister().RoleBindings(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	result := make([]*rbacv1.RoleBinding, 0)
	for _, rb := range rbs {
		if rb.RoleRef.Name == role {
			result = append(result, rb.DeepCopy())
		}
	}
	return result, nil
}

func (am *amOperator) ListWorkspaceRoles(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	conditions.Match[v1alpha2.OwnerName] = workspace
	conditions.Match[v1alpha2.OwnerKind] = "Workspace"
	result, err := am.resources.ListResources("", v1alpha2.ClusterRoles, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		return nil, err
	}

	for i, item := range result.Items {
		if role, ok := item.(*rbacv1.ClusterRole); ok {
			role = role.DeepCopy()
			role.Name = role.Annotations[constants.DisplayNameAnnotationKey]
			result.Items[i] = role
		}
	}
	return result, nil
}

func (am *amOperator) ListClusterRoles(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	return am.resources.ListResources("", v1alpha2.ClusterRoles, conditions, orderBy, reverse, limit, offset)
}

func (am *amOperator) NamespaceUsers(namespaceName string) ([]*User, error) {
	namespace, err := am.informers.Core().V1().Namespaces().Lister().Get(namespaceName)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	roleBindings, err := am.GetRoleBindings(namespaceName, "")

	if err != nil {
		return nil, err
	}

	users := make([]*User, 0)

	for _, roleBinding := range roleBindings {
		// controlled by ks-controller-manager
		if roleBinding.Name == NamespaceViewerRoleBindName {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {

				// show creator
				if roleBinding.Name == NamespaceAdminRoleBindName && subject.Name != namespace.Annotations[constants.CreatorAnnotationKey] {
					continue
				}

				user, err := GetUserInfo(subject.Name)

				if err != nil {
					if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
						continue
					}
					return nil, err
				}

				user.Role = roleBinding.RoleRef.Name
				user.RoleBindTime = &roleBinding.CreationTimestamp.Time
				user.RoleBinding = roleBinding.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (am *amOperator) GetUserWorkspaceSimpleRules(workspace, username string) ([]SimpleRule, error) {
	clusterRules, err := am.GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}

	// cluster-admin
	if RulesMatchesRequired(clusterRules, rbacv1.PolicyRule{
		Verbs:     []string{"*"},
		APIGroups: []string{"*"},
		Resources: []string{"*"},
	}) {
		return am.GetWorkspaceRoleSimpleRules(workspace, constants.WorkspaceAdmin), nil
	}

	workspaceRole, err := am.GetUserWorkspaceRole(workspace, username)

	if err != nil {
		if apierrors.IsNotFound(err) {

			// workspaces-manager
			if RulesMatchesRequired(clusterRules, rbacv1.PolicyRule{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"workspaces", "workspaces/*"},
			}) {
				return am.GetWorkspaceRoleSimpleRules(workspace, constants.WorkspacesManager), nil
			}

			return []SimpleRule{}, nil
		}

		klog.Error(err)
		return nil, err
	}

	return am.GetWorkspaceRoleSimpleRules(workspace, workspaceRole.Annotations[constants.DisplayNameAnnotationKey]), nil
}

func (am *amOperator) GetWorkspaceRoleSimpleRules(workspace, roleName string) []SimpleRule {

	workspaceRules := make([]SimpleRule, 0)

	switch roleName {
	case constants.WorkspaceAdmin:
		workspaceRules = []SimpleRule{
			{Name: "workspaces", Actions: []string{"edit", "delete", "view"}},
			{Name: "members", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "devops", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "projects", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "apps", Actions: []string{"view", "create", "manage"}},
			{Name: "repos", Actions: []string{"view", "manage"}},
		}
	case constants.WorkspaceRegular:
		workspaceRules = []SimpleRule{
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view", "create"}},
			{Name: "projects", Actions: []string{"view", "create"}},
			{Name: "apps", Actions: []string{"view", "create"}},
			{Name: "repos", Actions: []string{"view"}},
		}
	case constants.WorkspaceViewer:
		workspaceRules = []SimpleRule{
			{Name: "workspaces", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
			{Name: "projects", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "apps", Actions: []string{"view"}},
			{Name: "repos", Actions: []string{"view"}},
		}
	case constants.WorkspacesManager:
		workspaceRules = []SimpleRule{
			{Name: "workspaces", Actions: []string{"edit", "delete", "view"}},
			{Name: "members", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "roles", Actions: []string{"view"}},
		}
	}

	return workspaceRules
}

// Convert cluster role to rules
func (am *amOperator) GetClusterRoleSimpleRules(clusterRoleName string) ([]SimpleRule, error) {

	clusterRoleLister := am.informers.Rbac().V1().ClusterRoles().Lister()
	clusterRole, err := clusterRoleLister.Get(clusterRoleName)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return getClusterSimpleRule(clusterRole.Rules), nil
}

func (am *amOperator) GetUserClusterSimpleRules(username string) ([]SimpleRule, error) {
	clusterRules, err := am.GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}
	return getClusterSimpleRule(clusterRules), nil
}

func (am *amOperator) GetUserNamespaceSimpleRules(namespace, username string) ([]SimpleRule, error) {
	clusterRules, err := am.GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}
	rules, err := am.GetUserRules(namespace, username)
	if err != nil {
		return nil, err
	}
	rules = append(rules, clusterRules...)

	return getSimpleRule(rules), nil
}

// Convert roles to rules
func (am *amOperator) GetRoleSimpleRules(namespace string, roleName string) ([]SimpleRule, error) {

	roleLister := am.informers.Rbac().V1().Roles().Lister()
	role, err := roleLister.Roles(namespace).Get(roleName)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return getSimpleRule(role.Rules), nil
}

func getClusterSimpleRule(policyRules []rbacv1.PolicyRule) []SimpleRule {
	rules := make([]SimpleRule, 0)

	for i := 0; i < len(policy.ClusterRoleRuleMapping); i++ {
		validActions := make([]string, 0)
		for j := 0; j < (len(policy.ClusterRoleRuleMapping[i].Actions)); j++ {
			if rulesMatchesAction(policyRules, policy.ClusterRoleRuleMapping[i].Actions[j]) {
				validActions = append(validActions, policy.ClusterRoleRuleMapping[i].Actions[j].Name)
			}
		}
		if len(validActions) > 0 {
			rules = append(rules, SimpleRule{Name: policy.ClusterRoleRuleMapping[i].Name, Actions: validActions})
		}
	}

	return rules
}

func getSimpleRule(policyRules []rbacv1.PolicyRule) []SimpleRule {
	simpleRules := make([]SimpleRule, 0)
	for i := 0; i < len(policy.RoleRuleMapping); i++ {
		rule := SimpleRule{Name: policy.RoleRuleMapping[i].Name}
		rule.Actions = make([]string, 0)
		for j := 0; j < len(policy.RoleRuleMapping[i].Actions); j++ {
			if rulesMatchesAction(policyRules, policy.RoleRuleMapping[i].Actions[j]) {
				rule.Actions = append(rule.Actions, policy.RoleRuleMapping[i].Actions[j].Name)
			}
		}
		if len(rule.Actions) > 0 {
			simpleRules = append(simpleRules, rule)
		}
	}
	return simpleRules
}

func (am *amOperator) CreateClusterRoleBinding(username string, clusterRoleName string) error {
	clusterRoleLister := am.informers.Rbac().V1().ClusterRoles().Lister()

	_, err := clusterRoleLister.Get(clusterRoleName)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	if clusterRoleName == constants.ClusterAdmin {
		// create kubectl pod if cluster role is cluster-admin
		if err := kubectl.CreateKubectlDeploy(username); err != nil {
			klog.Error("create user terminal pod failed", username, err)
		}

	} else {
		// delete kubectl pod if cluster role is not cluster-admin, whether it exists or not
		if err := kubectl.DelKubectlDeploy(username); err != nil {
			klog.Error("delete user terminal pod failed", username, err)
		}
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	clusterRoleBinding.Name = username
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{Name: clusterRoleName, Kind: ClusterRoleKind}
	clusterRoleBinding.Subjects = []rbacv1.Subject{{Kind: rbacv1.UserKind, Name: username}}

	clusterRoleBindingLister := am.informers.Rbac().V1().ClusterRoleBindings().Lister()
	found, err := clusterRoleBindingLister.Get(username)

	if apierrors.IsNotFound(err) {
		_, err = client.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
		if err != nil {
			klog.Errorln("create cluster role binding", err)
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	// cluster role changed
	if found.RoleRef.Name != clusterRoleName {
		deletePolicy := metav1.DeletePropagationBackground
		gracePeriodSeconds := int64(0)
		deleteOption := &metav1.DeleteOptions{PropagationPolicy: &deletePolicy, GracePeriodSeconds: &gracePeriodSeconds}
		err = client.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Delete(found.Name, deleteOption)
		if err != nil {
			klog.Errorln(err)
			return err
		}
		_, err = client.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
		if err != nil {
			klog.Errorln(err)
			return err
		}
		return nil
	}

	if !k8sutil.ContainsUser(found.Subjects, username) {
		found.Subjects = clusterRoleBinding.Subjects
		_, err = client.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Update(found)
		if err != nil {
			klog.Errorln("update cluster role binding", err)
			return err
		}
	}

	return nil
}
