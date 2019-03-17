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
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"

	"k8s.io/api/rbac/v1"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/models"
)

const ClusterRoleKind = "ClusterRole"

// Get user list based on workspace role
func WorkspaceRoleUsers(workspace string, roleName string) ([]models.User, error) {

	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()

	workspaceRoleBinding, err := clusterRoleBindingLister.Get(fmt.Sprintf("system:%s:%s", workspace, roleName))

	if err != nil {
		return nil, err
	}

	names := make([]string, 0)

	for _, subject := range workspaceRoleBinding.Subjects {
		if subject.Kind == v1.UserKind {
			names = append(names, subject.Name)
		}
	}

	users, err := GetUsers(names)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(users); i++ {
		users[i].WorkspaceRole = roleName
	}

	return users, nil
}

func GetUsers(names []string) ([]models.User, error) {
	var users []models.User

	if names == nil || len(names) == 0 {
		return make([]models.User, 0), nil
	}

	conn, err := ldap.Client()

	if err != nil {
		return nil, err
	}

	for _, name := range names {
		user, err := UserDetail(name, conn)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func GetUser(name string) (*models.User, error) {

	conn, err := ldap.Client()

	if err != nil {
		return nil, err
	}

	user, err := UserDetail(name, conn)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserNamespaces(username string, requiredRule v1.PolicyRule) (allNamespace bool, namespaces []string, err error) {

	clusterRoles, err := GetClusterRoles(username)

	if err != nil {
		return false, nil, err
	}

	clusterRules := make([]v1.PolicyRule, 0)
	for _, role := range clusterRoles {
		clusterRules = append(clusterRules, role.Rules...)
	}

	if requiredRule.Size() == 0 {
		if RulesMatchesRequired(clusterRules, v1.PolicyRule{
			Verbs:     []string{"get"},
			APIGroups: []string{"kubesphere.io"},
			Resources: []string{"workspaces/namespaces"},
		}) {
			return true, nil, nil
		}
	} else {

		if RulesMatchesRequired(clusterRules, requiredRule) {
			return true, nil, nil
		}

	}

	roles, err := GetRoles("", username)

	if err != nil {
		return false, nil, err
	}

	rulesMapping := make(map[string][]v1.PolicyRule, 0)

	for _, role := range roles {
		rules := rulesMapping[role.Namespace]
		if rules == nil {
			rules = make([]v1.PolicyRule, 0)
		}
		rules = append(rules, role.Rules...)
		rulesMapping[role.Namespace] = rules
	}

	namespaces = make([]string, 0)

	for namespace, rules := range rulesMapping {
		if requiredRule.Size() == 0 || RulesMatchesRequired(rules, requiredRule) {
			namespaces = append(namespaces, namespace)
		}
	}

	return false, namespaces, nil
}

func GetWorkspaceUsers(workspace string, workspaceRole string) ([]string, error) {
	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBinding, err := clusterRoleBindingLister.Get(fmt.Sprintf("system:%s:%s", workspace, workspaceRole))

	if err != nil {
		return nil, err
	}

	users := make([]string, 0)

	for _, s := range clusterRoleBinding.Subjects {
		if s.Kind == v1.UserKind && !slice.ContainsString(users, s.Name, nil) {
			users = append(users, s.Name)
		}
	}
	return users, nil
}

func RulesMatchesRequired(rules []v1.PolicyRule, required v1.PolicyRule) bool {
	for _, rule := range rules {
		if ruleMatchesRequired(rule, required) {
			return true
		}
	}
	return false
}
