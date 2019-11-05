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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
	"kubesphere.io/kubesphere/pkg/models/resources"
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

func GetDevopsRoleSimpleRules(role string) []models.SimpleRule {
	var rules []models.SimpleRule

	switch role {
	case "developer":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"view", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "owner":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"edit", "view", "delete"}},
		}
		break
	case "maintainer":
		rules = []models.SimpleRule{
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
		rules = []models.SimpleRule{
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
func GetUserRoles(namespace, username string) ([]*rbacv1.Role, error) {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleLister := informers.SharedInformerFactory().Rbac().V1().Roles().Lister()
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln("get role bindings", namespace, err)
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
						klog.Errorln("get cluster role", roleBinding.RoleRef.Name, err)
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
						klog.Errorln("get role", roleBinding.Namespace, roleBinding.RoleRef.Name, err)
						return nil, err
					}
				}
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func GetUserClusterRoles(username string) (*rbacv1.ClusterRole, []*rbacv1.ClusterRole, error) {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		klog.Errorln("get cluster role bindings", err)
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
					klog.Errorln("get cluster role", clusterRoleBinding.RoleRef.Name, err)
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

func GetUserClusterRole(username string) (*rbacv1.ClusterRole, error) {
	userFacingClusterRole, _, err := GetUserClusterRoles(username)
	if err != nil {
		return nil, err
	}
	return userFacingClusterRole, nil
}

func GetUserClusterRules(username string) ([]rbacv1.PolicyRule, error) {
	_, clusterRoles, err := GetUserClusterRoles(username)

	if err != nil {
		return nil, err
	}

	rules := make([]rbacv1.PolicyRule, 0)
	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	return rules, nil
}

func GetUserRules(namespace, username string) ([]rbacv1.PolicyRule, error) {
	roles, err := GetUserRoles(namespace, username)

	if err != nil {
		return nil, err
	}

	rules := make([]rbacv1.PolicyRule, 0)
	for _, role := range roles {
		rules = append(rules, role.Rules...)
	}

	return rules, nil
}

func GetWorkspaceRoleBindings(workspace string) ([]*rbacv1.ClusterRoleBinding, error) {

	clusterRoleBindings, err := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

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

func GetWorkspaceRole(workspace, role string) (*rbacv1.ClusterRole, error) {
	if !sliceutil.HasString(constants.WorkSpaceRoles, role) {
		return nil, apierrors.NewNotFound(schema.GroupResource{Resource: "workspace role"}, role)
	}
	role = fmt.Sprintf("workspace:%s:%s", workspace, strings.TrimPrefix(role, "workspace-"))
	return informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister().Get(role)
}

func GetUserWorkspaceRoleMap(username string) (map[string]string, error) {

	clusterRoleBindings, err := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

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

func GetUserWorkspaceRole(workspace, username string) (*rbacv1.ClusterRole, error) {
	workspaceRoleMap, err := GetUserWorkspaceRoleMap(username)

	if err != nil {
		return nil, err
	}

	if workspaceRole := workspaceRoleMap[workspace]; workspaceRole != "" {
		return informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister().Get(workspaceRole)
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{Resource: "workspace user"}, username)
}

func GetRoleBindings(namespace string, roleName string) ([]*rbacv1.RoleBinding, error) {
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln("get role bindings", namespace, err)
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

func GetClusterRoleBindings(clusterRoleName string) ([]*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	roleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		klog.Errorln("get cluster role bindings", err)
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

func ListClusterRoleUsers(clusterRoleName string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	roleBindings, err := GetClusterRoleBindings(clusterRoleName)

	if err != nil {
		return nil, err
	}
	users := make([]*models.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				user, err := GetUserInfo(subject.Name)
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					klog.Errorln("get user info", subject.Name, err)
					return nil, err
				}
				users = append(users, user)
			}
		}
	}

	// order & reverse
	sort.Slice(users, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		switch orderBy {
		default:
			fallthrough
		case resources.Name:
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

func RoleUsers(namespace string, roleName string) ([]*models.User, error) {
	roleBindings, err := GetRoleBindings(namespace, roleName)

	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				user, err := GetUserInfo(subject.Name)

				if err != nil {
					if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
						continue
					}
					return nil, err
				}

				user.Role = roleBinding.RoleRef.Name

				users = append(users, user)
			}
		}
	}
	return users, nil
}

func ListRoles(namespace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	return resources.ListResources(namespace, resources.Roles, conditions, orderBy, reverse, limit, offset)
}

func ListWorkspaceRoles(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	conditions.Match[resources.OwnerName] = workspace
	conditions.Match[resources.OwnerKind] = "Workspace"
	result, err := resources.ListResources("", resources.ClusterRoles, conditions, orderBy, reverse, limit, offset)

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

func ListClusterRoles(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	return resources.ListResources("", resources.ClusterRoles, conditions, orderBy, reverse, limit, offset)
}

func NamespaceUsers(namespaceName string) ([]*models.User, error) {
	namespace, err := informers.SharedInformerFactory().Core().V1().Namespaces().Lister().Get(namespaceName)
	if err != nil {
		klog.Errorln("get namespace", namespaceName, err)
		return nil, err
	}
	roleBindings, err := GetRoleBindings(namespaceName, "")

	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0)

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

func GetUserWorkspaceSimpleRules(workspace, username string) ([]models.SimpleRule, error) {
	clusterRules, err := GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}

	// cluster-admin
	if RulesMatchesRequired(clusterRules, rbacv1.PolicyRule{
		Verbs:     []string{"*"},
		APIGroups: []string{"*"},
		Resources: []string{"*"},
	}) {
		return GetWorkspaceRoleSimpleRules(workspace, constants.WorkspaceAdmin), nil
	}

	workspaceRole, err := GetUserWorkspaceRole(workspace, username)

	if err != nil {
		if apierrors.IsNotFound(err) {

			// workspaces-manager
			if RulesMatchesRequired(clusterRules, rbacv1.PolicyRule{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"workspaces", "workspaces/*"},
			}) {
				return GetWorkspaceRoleSimpleRules(workspace, constants.WorkspacesManager), nil
			}

			return []models.SimpleRule{}, nil
		}

		klog.Error(err)
		return nil, err
	}

	return GetWorkspaceRoleSimpleRules(workspace, workspaceRole.Annotations[constants.DisplayNameAnnotationKey]), nil
}

func GetWorkspaceRoleSimpleRules(workspace, roleName string) []models.SimpleRule {

	workspaceRules := make([]models.SimpleRule, 0)

	switch roleName {
	case constants.WorkspaceAdmin:
		workspaceRules = []models.SimpleRule{
			{Name: "workspaces", Actions: []string{"edit", "delete", "view"}},
			{Name: "members", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "devops", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "projects", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "apps", Actions: []string{"view", "create", "manage"}},
			{Name: "repos", Actions: []string{"view", "manage"}},
		}
	case constants.WorkspaceRegular:
		workspaceRules = []models.SimpleRule{
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view", "create"}},
			{Name: "projects", Actions: []string{"view", "create"}},
			{Name: "apps", Actions: []string{"view", "create"}},
			{Name: "repos", Actions: []string{"view"}},
		}
	case constants.WorkspaceViewer:
		workspaceRules = []models.SimpleRule{
			{Name: "workspaces", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
			{Name: "projects", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "apps", Actions: []string{"view"}},
			{Name: "repos", Actions: []string{"view"}},
		}
	case constants.WorkspacesManager:
		workspaceRules = []models.SimpleRule{
			{Name: "workspaces", Actions: []string{"edit", "delete", "view"}},
			{Name: "members", Actions: []string{"edit", "delete", "create", "view"}},
			{Name: "roles", Actions: []string{"view"}},
		}
	}

	return workspaceRules
}

// Convert cluster role to rules
func GetClusterRoleSimpleRules(clusterRoleName string) ([]models.SimpleRule, error) {

	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	clusterRole, err := clusterRoleLister.Get(clusterRoleName)

	if err != nil {
		klog.Errorln("get cluster role", clusterRoleName, clusterRoleName)
		return nil, err
	}

	return getClusterSimpleRule(clusterRole.Rules), nil
}

func GetUserClusterSimpleRules(username string) ([]models.SimpleRule, error) {
	clusterRules, err := GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}
	return getClusterSimpleRule(clusterRules), nil
}

func GetUserNamespaceSimpleRules(namespace, username string) ([]models.SimpleRule, error) {
	clusterRules, err := GetUserClusterRules(username)
	if err != nil {
		return nil, err
	}
	rules, err := GetUserRules(namespace, username)
	if err != nil {
		return nil, err
	}
	rules = append(rules, clusterRules...)

	return getSimpleRule(rules), nil
}

// Convert roles to rules
func GetRoleSimpleRules(namespace string, roleName string) ([]models.SimpleRule, error) {

	roleLister := informers.SharedInformerFactory().Rbac().V1().Roles().Lister()
	role, err := roleLister.Roles(namespace).Get(roleName)

	if err != nil {
		klog.Errorln("get role", namespace, roleName, err)
		return nil, err
	}

	return getSimpleRule(role.Rules), nil
}

func getClusterSimpleRule(policyRules []rbacv1.PolicyRule) []models.SimpleRule {
	rules := make([]models.SimpleRule, 0)

	for i := 0; i < len(policy.ClusterRoleRuleMapping); i++ {
		validActions := make([]string, 0)
		for j := 0; j < (len(policy.ClusterRoleRuleMapping[i].Actions)); j++ {
			if rulesMatchesAction(policyRules, policy.ClusterRoleRuleMapping[i].Actions[j]) {
				validActions = append(validActions, policy.ClusterRoleRuleMapping[i].Actions[j].Name)
			}
		}
		if len(validActions) > 0 {
			rules = append(rules, models.SimpleRule{Name: policy.ClusterRoleRuleMapping[i].Name, Actions: validActions})
		}
	}

	return rules
}

func getSimpleRule(policyRules []rbacv1.PolicyRule) []models.SimpleRule {
	simpleRules := make([]models.SimpleRule, 0)
	for i := 0; i < len(policy.RoleRuleMapping); i++ {
		rule := models.SimpleRule{Name: policy.RoleRuleMapping[i].Name}
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

func CreateClusterRoleBinding(username string, clusterRoleName string) error {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()

	_, err := clusterRoleLister.Get(clusterRoleName)

	if err != nil {
		klog.Errorln("get cluster role", clusterRoleName, err)
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

	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
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

func RulesMatchesRequired(rules []rbacv1.PolicyRule, required rbacv1.PolicyRule) bool {
	for _, rule := range rules {
		if ruleMatchesRequired(rule, required) {
			return true
		}
	}
	return false
}

func rulesMatchesAction(rules []rbacv1.PolicyRule, action models.Action) bool {

	for _, required := range action.Rules {
		if !RulesMatchesRequired(rules, required) {
			return false
		}
	}

	return true
}

func ruleMatchesRequired(rule rbacv1.PolicyRule, required rbacv1.PolicyRule) bool {

	if len(required.NonResourceURLs) == 0 {
		for _, apiGroup := range required.APIGroups {
			for _, resource := range required.Resources {
				resources := strings.Split(resource, "/")
				resource = resources[0]
				var subsource string
				if len(resources) > 1 {
					subsource = resources[1]
				}

				if len(required.ResourceNames) == 0 {
					for _, verb := range required.Verbs {
						if !ruleMatchesRequest(rule, apiGroup, "", resource, subsource, "", verb) {
							return false
						}
					}
				} else {
					for _, resourceName := range required.ResourceNames {
						for _, verb := range required.Verbs {
							if !ruleMatchesRequest(rule, apiGroup, "", resource, subsource, resourceName, verb) {
								return false
							}
						}
					}
				}
			}
		}
	} else {
		for _, apiGroup := range required.APIGroups {
			for _, nonResourceURL := range required.NonResourceURLs {
				for _, verb := range required.Verbs {
					if !ruleMatchesRequest(rule, apiGroup, nonResourceURL, "", "", "", verb) {
						return false
					}
				}
			}
		}
	}
	return true
}

func ruleMatchesResources(rule rbacv1.PolicyRule, apiGroup string, resource string, subresource string, resourceName string) bool {

	if resource == "" {
		return false
	}

	if !hasString(rule.APIGroups, apiGroup) && !hasString(rule.APIGroups, rbacv1.ResourceAll) {
		return false
	}

	if len(rule.ResourceNames) > 0 && !hasString(rule.ResourceNames, resourceName) {
		return false
	}

	combinedResource := resource

	if subresource != "" {
		combinedResource = combinedResource + "/" + subresource
	}

	for _, res := range rule.Resources {

		// match "*"
		if res == rbacv1.ResourceAll || res == combinedResource {
			return true
		}

		// match "*/subresource"
		if len(subresource) > 0 && strings.HasPrefix(res, "*/") && subresource == strings.TrimLeft(res, "*/") {
			return true
		}
		// match "resource/*"
		if strings.HasSuffix(res, "/*") && resource == strings.TrimRight(res, "/*") {
			return true
		}
	}

	return false
}

func ruleMatchesRequest(rule rbacv1.PolicyRule, apiGroup string, nonResourceURL string, resource string, subresource string, resourceName string, verb string) bool {

	if !hasString(rule.Verbs, verb) && !hasString(rule.Verbs, rbacv1.VerbAll) {
		return false
	}

	if nonResourceURL == "" {
		return ruleMatchesResources(rule, apiGroup, resource, subresource, resourceName)
	} else {
		return ruleMatchesNonResource(rule, nonResourceURL)
	}
}

func ruleMatchesNonResource(rule rbacv1.PolicyRule, nonResourceURL string) bool {

	if nonResourceURL == "" {
		return false
	}

	for _, spec := range rule.NonResourceURLs {
		if pathMatches(nonResourceURL, spec) {
			return true
		}
	}

	return false
}

func pathMatches(path, spec string) bool {
	// Allow wildcard match
	if spec == "*" {
		return true
	}
	// Allow exact match
	if spec == path {
		return true
	}
	// Allow a trailing * subpath match
	if strings.HasSuffix(spec, "*") && strings.HasPrefix(path, strings.TrimRight(spec, "*")) {
		return true
	}
	return false
}

func hasString(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}
