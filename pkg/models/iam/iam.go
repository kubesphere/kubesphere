package iam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	v12 "k8s.io/client-go/listers/rbac/v1"

	"kubesphere.io/kubesphere/pkg/informers"

	"k8s.io/api/rbac/v1"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/constants"
	ksErr "kubesphere.io/kubesphere/pkg/errors"
	. "kubesphere.io/kubesphere/pkg/models"
)

const ClusterRoleKind = "ClusterRole"

var (
	clusterRoleBindingLister v12.ClusterRoleBindingLister
	clusterRoleLister        v12.ClusterRoleLister
	roleBindingLister        v12.RoleBindingLister
	roleLister               v12.RoleLister
)

func init() {
	clusterRoleBindingLister = informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleLister = informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister()
	roleBindingLister = informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleLister = informers.SharedInformerFactory().Rbac().V1().Roles().Lister()
}

// Get user list based on workspace role
func WorkspaceRoleUsers(workspace string, roleName string) ([]User, error) {

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

func GetUsers(names []string) ([]User, error) {
	var users []User

	if names == nil || len(names) == 0 {
		return make([]User, 0), nil
	}

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/users?name=%s", constants.AccountAPIServer, strings.Join(names, ",")))

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	err = json.Unmarshal(data, &users)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUser(name string) (*User, error) {

	result, err := http.Get(fmt.Sprintf("http://%s/apis/account.kubesphere.io/v1alpha1/users/%s", constants.AccountAPIServer, name))

	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	if result.StatusCode > 200 {
		return nil, ksErr.Wrap(data)
	}

	var user User

	err = json.Unmarshal(data, &user)

	if err != nil {
		return nil, err
	}

	return &user, nil
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
