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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-ldap/ldap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
)

func GetNamespaces(username string) ([]*corev1.Namespace, error) {

	roles, err := GetRoles(username, "")

	if err != nil {
		return nil, err
	}

	namespaces := make([]*corev1.Namespace, 0)
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	for _, role := range roles {
		namespace, err := namespaceLister.Get(role.Name)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func GetNamespacesByWorkspace(workspace string) ([]*corev1.Namespace, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	return namespaceLister.List(labels.SelectorFromSet(labels.Set{"kubesphere.io/workspace": workspace}))
}

func GetDevopsRole(projectId string, username string) (string, error) {

	//Hard fix
	if username == "admin" {
		return "owner", nil
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/api/v1alpha/projects/%s/members", constants.DevopsAPIServer, projectId), nil)

	if err != nil {
		return "", err
	}
	req.Header.Set(constants.UserNameHeader, username)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode > 200 {
		return "", errors.New(string(data))
	}

	var result []map[string]string

	err = json.Unmarshal(data, &result)

	if err != nil {
		return "", err
	}

	for _, item := range result {
		if item["username"] == username {
			return item["role"], nil
		}
	}

	return "", nil
}

func GetNamespace(namespaceName string) (*corev1.Namespace, error) {
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	return namespaceLister.Get(namespaceName)
}

func GetRoles(username string, namespace string) ([]*v1.Role, error) {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleLister := informers.SharedInformerFactory().Rbac().V1().Roles().Lister()
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	roles := make([]*v1.Role, 0)

	for _, roleBinding := range roleBindings {

		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {
					clusterRole, err := clusterRoleLister.Get(roleBinding.RoleRef.Name)
					if err == nil {
						var role = v1.Role{TypeMeta: (*clusterRole).TypeMeta, ObjectMeta: (*clusterRole).ObjectMeta, Rules: (*clusterRole).Rules}
						role.Namespace = roleBinding.Namespace
						roles = append(roles, &role)
						break
					} else if apierrors.IsNotFound(err) {
						log.Println(err)
						break
					} else {
						return nil, err
					}
				} else {
					if subject.Kind == v1.UserKind && subject.Name == username {
						rule, err := roleLister.Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name)
						if err == nil {
							roles = append(roles, rule)
							break
						} else if apierrors.IsNotFound(err) {
							log.Println(err)
							break
						} else {
							return nil, err
						}

					}
				}
			}
		}
	}

	return roles, nil
}

func GetClusterRoles(username string) ([]*v1.ClusterRole, error) {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.SelectorFromSet(labels.Set{"": ""}))

	if err != nil {
		return nil, err
	}

	roles := make([]*v1.ClusterRole, 0)

	for _, rb := range clusterRoleBindings {
		if rb.RoleRef.Kind == ClusterRoleKind {
			for _, subject := range rb.Subjects {
				if subject.Kind == v1.UserKind && subject.Name == username {

					role, err := clusterRoleLister.Get(rb.RoleRef.Name)
					role = role.DeepCopy()
					if err == nil {
						if role.Annotations == nil {
							role.Annotations = make(map[string]string, 0)
						}

						role.Annotations["rbac.authorization.k8s.io/clusterrolebinding"] = rb.Name

						if rb.Annotations != nil &&
							rb.Annotations["rbac.authorization.k8s.io/clusterrole"] == rb.RoleRef.Name {
							role.Annotations["rbac.authorization.k8s.io/clusterrole"] = "true"
						}

						roles = append(roles, role)
						break
					} else if apierrors.IsNotFound(err) {
						glog.Warningln(err)
						break
					} else {
						return nil, err
					}
				}
			}
		}
	}

	return roles, nil
}

func GetRoleBindings(namespace string, roleName string) ([]*v1.RoleBinding, error) {
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleBindingList, err := roleBindingLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	items := make([]*v1.RoleBinding, 0)

	for _, roleBinding := range roleBindingList {
		if roleName == "" {
			items = append(items, roleBinding)
		} else if roleBinding.RoleRef.Name == roleName {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func GetClusterRoleBindings(clusterRoleName string) ([]*v1.ClusterRoleBinding, error) {
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	roleBindingList, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	items := make([]*v1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindingList {
		if roleBinding.RoleRef.Name == clusterRoleName {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func ClusterRoleUsers(clusterRoleName string) ([]*models.User, error) {

	roleBindings, err := GetClusterRoleBindings(clusterRoleName)

	if err != nil {
		return nil, err
	}

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	names := make([]string, 0)
	users := make([]*models.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && !strings.HasPrefix(subject.Name, "system") &&
				!slice.ContainsString(names, subject.Name, nil) {
				names = append(names, subject.Name)

				user, err := UserDetail(subject.Name, conn)

				if ldap.IsErrorWithCode(err, 32) {
					continue
				}

				if err != nil {
					return nil, err
				}

				users = append(users, user)
			}
		}
	}

	return users, nil

}

func RoleUsers(namespace string, roleName string) ([]*models.User, error) {
	roleBindings, err := GetRoleBindings(namespace, roleName)

	if err != nil {
		return nil, err
	}

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	names := make([]string, 0)
	users := make([]*models.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind &&
				!strings.HasPrefix(subject.Name, "system") &&
				!slice.ContainsString(names, subject.Name, nil) {
				names = append(names, subject.Name)
				user, err := UserDetail(subject.Name, conn)
				if ldap.IsErrorWithCode(err, 32) {
					continue
				}

				if err != nil {
					return nil, err
				}

				users = append(users, user)
			}
		}
	}
	return users, nil
}

func NamespaceUsers(namespaceName string) ([]*models.User, error) {
	roleBindings, err := GetRoleBindings(namespaceName, "")
	if err != nil {
		return nil, err
	}
	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	names := make([]string, 0)
	users := make([]*models.User, 0)

	for _, roleBinding := range roleBindings {

		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind &&
				!slice.ContainsString(names, subject.Name, nil) &&
				!strings.HasPrefix(subject.Name, "system") {
				if roleBinding.Name == "viewer" {
					continue
				}
				if roleBinding.Name == "admin" {
					continue
				}
				names = append(names, subject.Name)
				user, err := UserDetail(subject.Name, conn)
				if ldap.IsErrorWithCode(err, 32) {
					continue
				}
				if err != nil {
					return nil, err
				}
				user.Role = roleBinding.RoleRef.Name
				user.RoleBinding = roleBinding.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func GetWorkspaceRoles(clusterRoles []*v1.ClusterRole) map[string]string {

	workspaceRoles := make(map[string]string, 0)

	for _, v := range clusterRoles {
		if groups := regexp.MustCompile(fmt.Sprintf(`^system:(\S+):(%s)$`, strings.Join(constants.WorkSpaceRoles, "|"))).FindStringSubmatch(v.Name); len(groups) == 3 {
			workspaceRoles[groups[1]] = groups[2]
		}
	}

	return workspaceRoles
}

func GetWorkspaceRole(clusterRoles []*v1.ClusterRole, workspace string) string {

	for _, v := range clusterRoles {
		if groups := regexp.MustCompile(fmt.Sprintf(`^system:(\S+):(%s)$`, strings.Join(constants.WorkSpaceRoles, "|"))).FindStringSubmatch(v.Name); len(groups) == 3 {
			if groups[1] == workspace {
				return groups[2]
			}
		}
	}

	return ""
}

func GetWorkspaceSimpleRules(clusterRoles []*v1.ClusterRole, workspace string) map[string][]models.SimpleRule {

	workspaceRules := make(map[string][]models.SimpleRule, 0)

	clusterSimpleRules := make([]models.SimpleRule, 0)
	clusterRules := make([]v1.PolicyRule, 0)
	for _, clusterRole := range clusterRoles {
		clusterRules = append(clusterRules, clusterRole.Rules...)
	}

	for i := 0; i < len(policy.WorkspaceRoleRuleMapping); i++ {
		rule := models.SimpleRule{Name: policy.WorkspaceRoleRuleMapping[i].Name}
		rule.Actions = make([]string, 0)
		for j := 0; j < (len(policy.WorkspaceRoleRuleMapping[i].Actions)); j++ {
			if RulesMatchesAction(clusterRules, policy.WorkspaceRoleRuleMapping[i].Actions[j]) {
				rule.Actions = append(rule.Actions, policy.WorkspaceRoleRuleMapping[i].Actions[j].Name)
			}
		}
		if len(rule.Actions) > 0 {
			clusterSimpleRules = append(clusterSimpleRules, rule)
		}
	}

	if len(clusterRules) > 0 {
		workspaceRules["*"] = clusterSimpleRules
	}

	for _, v := range clusterRoles {

		if groups := regexp.MustCompile(fmt.Sprintf(`^system:(\S+):(%s)$`, strings.Join(constants.WorkSpaceRoles, "|"))).FindStringSubmatch(v.Name); len(groups) == 3 {

			if workspace != "" && groups[1] != workspace {
				continue
			}

			policyRules := make([]v1.PolicyRule, 0)

			for _, rule := range v.Rules {
				rule.ResourceNames = nil
				policyRules = append(policyRules, rule)
			}

			rules := make([]models.SimpleRule, 0)

			for i := 0; i < len(policy.WorkspaceRoleRuleMapping); i++ {
				rule := models.SimpleRule{Name: policy.WorkspaceRoleRuleMapping[i].Name}
				rule.Actions = make([]string, 0)
				for j := 0; j < (len(policy.WorkspaceRoleRuleMapping[i].Actions)); j++ {
					action := policy.WorkspaceRoleRuleMapping[i].Actions[j]
					if RulesMatchesAction(policyRules, action) {
						rule.Actions = append(rule.Actions, action.Name)
					}
				}
				if len(rule.Actions) > 0 {
					rules = append(rules, rule)
				}
			}

			workspaceRules[groups[1]] = merge(rules, clusterSimpleRules)
		}
	}

	return workspaceRules
}

func merge(clusterRules, rules []models.SimpleRule) []models.SimpleRule {
	for _, clusterRule := range clusterRules {
		exist := false

		for i := 0; i < len(rules); i++ {
			if rules[i].Name == clusterRule.Name {
				exist = true

				for _, action := range clusterRule.Actions {
					if !slice.ContainsString(rules[i].Actions, action, nil) {
						rules[i].Actions = append(rules[i].Actions, action)
					}
				}
			}
		}

		if !exist {
			rules = append(rules, clusterRule)
		}
	}
	return rules
}

// Convert cluster roles to rules
func GetClusterRoleSimpleRules(clusterRoles []*v1.ClusterRole) ([]models.SimpleRule, error) {

	clusterRules := make([]v1.PolicyRule, 0)

	for _, v := range clusterRoles {
		clusterRules = append(clusterRules, v.Rules...)
	}

	rules := make([]models.SimpleRule, 0)

	for i := 0; i < len(policy.ClusterRoleRuleMapping); i++ {
		validActions := make([]string, 0)
		for j := 0; j < (len(policy.ClusterRoleRuleMapping[i].Actions)); j++ {
			if RulesMatchesAction(clusterRules, policy.ClusterRoleRuleMapping[i].Actions[j]) {
				validActions = append(validActions, policy.ClusterRoleRuleMapping[i].Actions[j].Name)
			}
		}
		if len(validActions) > 0 {
			rules = append(rules, models.SimpleRule{Name: policy.ClusterRoleRuleMapping[i].Name, Actions: validActions})
		}
	}

	return rules, nil
}

// Convert roles to rules
func GetRoleSimpleRules(roles []*v1.Role, namespace string) (map[string][]models.SimpleRule, error) {

	rulesMapping := make(map[string][]models.SimpleRule, 0)

	policyRulesMapping := make(map[string][]v1.PolicyRule, 0)

	for _, v := range roles {

		if namespace != "" && v.Namespace != namespace {
			continue
		}

		policyRules := policyRulesMapping[v.Namespace]

		if policyRules == nil {
			policyRules = make([]v1.PolicyRule, 0)
		}

		policyRules = append(policyRules, v.Rules...)

		policyRulesMapping[v.Namespace] = policyRules
	}

	for namespace, policyRules := range policyRulesMapping {

		rules := make([]models.SimpleRule, 0)

		for i := 0; i < len(policy.RoleRuleMapping); i++ {
			rule := models.SimpleRule{Name: policy.RoleRuleMapping[i].Name}
			rule.Actions = make([]string, 0)
			for j := 0; j < len(policy.RoleRuleMapping[i].Actions); j++ {
				if RulesMatchesAction(policyRules, policy.RoleRuleMapping[i].Actions[j]) {
					rule.Actions = append(rule.Actions, policy.RoleRuleMapping[i].Actions[j].Name)
				}
			}
			if len(rule.Actions) > 0 {
				rules = append(rules, rule)
			}
		}

		rulesMapping[namespace] = rules
	}

	return rulesMapping, nil
}

//
func CreateClusterRoleBinding(username string, clusterRoleName string) error {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()

	_, err := clusterRoleLister.Get(clusterRoleName)

	if err != nil {
		return err
	}

	clusterRoles, err := GetClusterRoles(username)

	if err != nil {
		return err
	}

	for _, clusterRole := range clusterRoles {

		if clusterRole.Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {

			if clusterRole.Name == clusterRoleName {
				return nil
			}

			clusterRoleBindingName := clusterRole.Annotations["rbac.authorization.k8s.io/clusterrolebinding"]
			clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
			clusterRoleBinding, err := clusterRoleBindingLister.Get(clusterRoleBindingName)

			if err != nil {
				return err
			}

			for i, v := range clusterRoleBinding.Subjects {
				if v.Kind == v1.UserKind && v.Name == username {
					clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects[:i], clusterRoleBinding.Subjects[i+1:]...)
					break
				}
			}

			_, err = k8s.Client().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)

			if err != nil {
				return err
			}

			break
		}
	}
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		return err
	}

	var clusterRoleBinding *v1.ClusterRoleBinding

	for _, roleBinding := range clusterRoleBindings {
		if roleBinding.Annotations != nil && roleBinding.Annotations["rbac.authorization.k8s.io/clusterrole"] == clusterRoleName &&
			roleBinding.RoleRef.Name == clusterRoleName {
			clusterRoleBinding = roleBinding
			break
		}
	}

	if clusterRoleBinding != nil {
		clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, v1.Subject{Kind: v1.UserKind, Name: username})
		_, err := k8s.Client().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
		if err != nil {
			return err
		}
	} else {
		clusterRoleBinding = new(v1.ClusterRoleBinding)
		clusterRoleBinding.Annotations = map[string]string{"rbac.authorization.k8s.io/clusterrole": clusterRoleName}
		clusterRoleBinding.Name = clusterRoleName
		clusterRoleBinding.RoleRef = v1.RoleRef{Name: clusterRoleName, Kind: ClusterRoleKind}
		clusterRoleBinding.Subjects = []v1.Subject{{Kind: v1.UserKind, Name: username}}

		_, err = k8s.Client().RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)

		if err != nil {
			return err
		}
	}

	return nil
}

func GetRole(namespace string, roleName string) (*v1.Role, error) {
	return informers.SharedInformerFactory().Rbac().V1().Roles().Lister().Roles(namespace).Get(roleName)
}
func GetClusterRole(clusterRoleName string) (*v1.ClusterRole, error) {
	clusterRoleLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	return clusterRoleLister.Get(clusterRoleName)
}

func RulesMatchesAction(rules []v1.PolicyRule, action models.Action) bool {

	for _, required := range action.Rules {
		if !rulesMatchesRequired(rules, required) {
			return false
		}
	}

	return true
}

func rulesMatchesRequired(rules []v1.PolicyRule, required v1.PolicyRule) bool {
	for _, rule := range rules {
		if ruleMatchesRequired(rule, required) {
			return true
		}
	}
	return false
}

func ruleMatchesRequired(rule v1.PolicyRule, required v1.PolicyRule) bool {

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

func ruleMatchesResources(rule v1.PolicyRule, apiGroup string, resource string, subresource string, resourceName string) bool {

	if resource == "" {
		return false
	}

	if !hasString(rule.APIGroups, apiGroup) && !hasString(rule.APIGroups, v1.ResourceAll) {
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
		if res == v1.ResourceAll || res == combinedResource {
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

func ruleMatchesRequest(rule v1.PolicyRule, apiGroup string, nonResourceURL string, resource string, subresource string, resourceName string, verb string) bool {

	if !hasString(rule.Verbs, verb) && !hasString(rule.Verbs, v1.VerbAll) {
		return false
	}

	if nonResourceURL == "" {
		return ruleMatchesResources(rule, apiGroup, resource, subresource, resourceName)
	} else {
		return ruleMatchesNonResource(rule, nonResourceURL)
	}
}

func ruleMatchesNonResource(rule v1.PolicyRule, nonResourceURL string) bool {

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
