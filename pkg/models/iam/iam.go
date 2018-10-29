package iam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	ksErr "kubesphere.io/kubesphere/pkg/util/errors"
)

const ClusterRoleKind = "ClusterRole"

// Get user list based on workspace role
func WorkspaceRoleUsers(workspace string, roleName string) ([]User, error) {

	k8sClient := client.NewK8sClient()

	roleBinding, err := k8sClient.RbacV1().ClusterRoleBindings().Get(fmt.Sprintf("system:%s:%s", workspace, roleName), meta_v1.GetOptions{})

	if err != nil {
		return nil, err
	}

	names := make([]string, 0)

	for _, subject := range roleBinding.Subjects {
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

// Get rules
func WorkspaceRoleRules(workspace string, roleName string) (*v1.ClusterRole, []Rule, error) {
	k8sClient := client.NewK8sClient()

	role, err := k8sClient.RbacV1().ClusterRoles().Get(fmt.Sprintf("system:%s:%s", workspace, roleName), meta_v1.GetOptions{})

	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(role.Rules); i++ {
		role.Rules[i].ResourceNames = nil
	}

	rules := make([]Rule, 0)
	for i := 0; i < len(WorkspaceRoleRuleMapping); i++ {
		rule := Rule{Name: WorkspaceRoleRuleMapping[i].Name}
		rule.Actions = make([]Action, 0)
		for j := 0; j < len(WorkspaceRoleRuleMapping[i].Actions); j++ {
			if actionValidate(role.Rules, WorkspaceRoleRuleMapping[i].Actions[j]) {
				rule.Actions = append(rule.Actions, WorkspaceRoleRuleMapping[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	role.Name = roleName

	return role, rules, nil
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
		if ruleValidate(clusterRules, v1.PolicyRule{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
		}) {
			return true, nil, nil
		}
	} else if ruleValidate(clusterRules, requiredRule) {
		return true, nil, nil
	}

	roles, err := GetRoles(username)

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
		if requiredRule.Size() == 0 || ruleValidate(rules, requiredRule) {
			namespaces = append(namespaces, namespace)
		}
	}

	return false, namespaces, nil
}

func DeleteRoleBindings(username string) error {
	k8s := client.NewK8sClient()

	roleBindings, err := k8s.RbacV1().RoleBindings("").List(meta_v1.ListOptions{})

	if err != nil {
		return err
	}

	for _, roleBinding := range roleBindings.Items {

		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)

		if length2 == 0 {
			k8s.RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &meta_v1.DeleteOptions{})
		} else if length2 < length1 {
			k8s.RbacV1().RoleBindings(roleBinding.Namespace).Update(&roleBinding)
		}
	}

	clusterRoleBindingList, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	for _, roleBinding := range clusterRoleBindingList.Items {
		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)
		if length2 == 0 {
			k8s.RbacV1().ClusterRoleBindings().Delete(roleBinding.Name, &meta_v1.DeleteOptions{})
		} else if length2 < length1 {
			k8s.RbacV1().ClusterRoleBindings().Update(&roleBinding)
		}
	}

	return nil
}

func GetRole(namespace string, name string) (*v1.Role, error) {
	k8s := client.NewK8sClient()
	role, err := k8s.RbacV1().Roles(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func GetClusterRoleBindings(name string) ([]v1.ClusterRoleBinding, error) {
	k8s := client.NewK8sClient()
	roleBindingList, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	items := make([]v1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.Name == name {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func GetRoleBindings(namespace string, name string) ([]v1.RoleBinding, error) {
	k8s := client.NewK8sClient()

	roleBindingList, err := k8s.RbacV1().RoleBindings(namespace).List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	items := make([]v1.RoleBinding, 0)

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.Name == name {
			items = append(items, roleBinding)
		}
	}

	return items, nil
}

func GetClusterRole(name string) (*v1.ClusterRole, error) {
	k8s := client.NewK8sClient()
	role, err := k8s.RbacV1().ClusterRoles().Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func GetRoles(username string) ([]v1.Role, error) {
	k8s := client.NewK8sClient()

	roleBindings, err := k8s.RbacV1().RoleBindings("").List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}
	roles := make([]v1.Role, 0)

	for _, roleBinding := range roleBindings.Items {

		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {
					clusterRole, err := k8s.RbacV1().ClusterRoles().Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})
					if err == nil {
						var role = v1.Role{TypeMeta: (*clusterRole).TypeMeta, ObjectMeta: (*clusterRole).ObjectMeta, Rules: (*clusterRole).Rules}
						role.Namespace = roleBinding.Namespace
						roles = append(roles, role)
						break
					} else if apierrors.IsNotFound(err) {
						glog.Infoln(err.Error())
						break
					} else {
						return nil, err
					}

				} else {
					if subject.Kind == v1.UserKind && subject.Name == username {
						rule, err := k8s.RbacV1().Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})
						if err == nil {
							roles = append(roles, *rule)
							break
						} else if apierrors.IsNotFound(err) {
							glog.Infoln(err.Error())
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

func GetClusterRoles(username string) ([]v1.ClusterRole, error) {
	k8s := client.NewK8sClient()

	clusterRoleBindings, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	roles := make([]v1.ClusterRole, 0)

	for _, roleBinding := range clusterRoleBindings.Items {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {
					role, err := k8s.RbacV1().ClusterRoles().Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})
					if err == nil {
						if role.Annotations == nil {
							role.Annotations = make(map[string]string, 0)
						}

						role.Annotations["rbac.authorization.k8s.io/clusterrolebinding"] = roleBinding.Name

						if roleBinding.Annotations != nil &&
							roleBinding.Annotations["rbac.authorization.k8s.io/clusterrole"] == roleBinding.RoleRef.Name {
							role.Annotations["rbac.authorization.k8s.io/clusterrole"] = "true"
						}

						roles = append(roles, *role)
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

	return roles, nil
}

func ruleValidate(rules []v1.PolicyRule, rule v1.PolicyRule) bool {

	for _, apiGroup := range rule.APIGroups {
		if len(rule.NonResourceURLs) == 0 {
			for _, resource := range rule.Resources {

				//if len(Rule.ResourceNames) == 0 {

				for _, verb := range rule.Verbs {
					if !verbValidate(rules, apiGroup, "", resource, "", verb) {
						return false
					}
				}

				//} else {
				//	for _, resourceName := range Rule.ResourceNames {
				//		for _, verb := range Rule.Verbs {
				//			if !verbValidate(rules, apiGroup, "", resource, resourceName, verb) {
				//				return false
				//			}
				//		}
				//	}
				//}
			}
		} else {
			for _, nonResourceURL := range rule.NonResourceURLs {
				for _, verb := range rule.Verbs {
					if !verbValidate(rules, apiGroup, nonResourceURL, "", "", verb) {
						return false
					}
				}
			}
		}
	}
	return true
}

func verbValidate(rules []v1.PolicyRule, apiGroup string, nonResourceURL string, resource string, resourceName string, verb string) bool {
	for _, rule := range rules {

		if nonResourceURL == "" {
			if slice.ContainsString(rule.APIGroups, apiGroup, nil) ||
				slice.ContainsString(rule.APIGroups, v1.APIGroupAll, nil) {
				if slice.ContainsString(rule.Verbs, verb, nil) ||
					slice.ContainsString(rule.Verbs, v1.VerbAll, nil) {
					if slice.ContainsString(rule.Resources, v1.ResourceAll, nil) {
						return true
					} else if slice.ContainsString(rule.Resources, resource, nil) {
						if len(rule.ResourceNames) > 0 {
							if slice.ContainsString(rule.ResourceNames, resourceName, nil) {
								return true
							}
						} else if resourceName == "" {
							return true
						}
					}
				}
			}

		} else if slice.ContainsString(rule.NonResourceURLs, nonResourceURL, nil) ||
			slice.ContainsString(rule.NonResourceURLs, v1.NonResourceAll, nil) {
			if slice.ContainsString(rule.Verbs, verb, nil) ||
				slice.ContainsString(rule.Verbs, v1.VerbAll, nil) {
				return true
			}
		}
	}
	return false
}

func GetUserRules(username string) (map[string][]Rule, error) {

	items := make(map[string][]Rule, 0)
	userRoles, err := GetRoles(username)

	if err != nil {
		return nil, err
	}

	rulesMapping := make(map[string][]v1.PolicyRule, 0)

	for _, role := range userRoles {
		rules := rulesMapping[role.Namespace]
		if rules == nil {
			rules = make([]v1.PolicyRule, 0)
		}
		rules = append(rules, role.Rules...)
		rulesMapping[role.Namespace] = rules
	}

	for namespace, policyRules := range rulesMapping {
		rules := convertToRules(policyRules)
		if len(rules) > 0 {
			items[namespace] = rules
		}
	}

	return items, nil
}

func convertToRules(policyRules []v1.PolicyRule) []Rule {
	rules := make([]Rule, 0)

	for i := 0; i < (len(RoleRuleGroup)); i++ {
		rule := Rule{Name: RoleRuleGroup[i].Name}
		rule.Actions = make([]Action, 0)
		for j := 0; j < (len(RoleRuleGroup[i].Actions)); j++ {
			if actionValidate(policyRules, RoleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, RoleRuleGroup[i].Actions[j])
			}
		}

		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules
}

func GetUserClusterRules(username string) ([]Rule, error) {

	rules := make([]Rule, 0)

	clusterRoles, err := GetClusterRoles(username)

	if err != nil {
		return nil, err
	}

	clusterRules := make([]v1.PolicyRule, 0)

	for _, role := range clusterRoles {
		clusterRules = append(clusterRules, role.Rules...)
	}

	for i := 0; i < (len(ClusterRoleRuleGroup)); i++ {
		rule := Rule{Name: ClusterRoleRuleGroup[i].Name}
		rule.Actions = make([]Action, 0)
		for j := 0; j < (len(ClusterRoleRuleGroup[i].Actions)); j++ {
			if actionValidate(clusterRules, ClusterRoleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, ClusterRoleRuleGroup[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

func GetClusterRoleRules(name string) ([]Rule, error) {

	clusterRole, err := GetClusterRole(name)

	if err != nil {
		return nil, err
	}

	rules := make([]Rule, 0)

	for i := 0; i < len(ClusterRoleRuleGroup); i++ {
		rule := Rule{Name: ClusterRoleRuleGroup[i].Name}
		rule.Actions = make([]Action, 0)
		for j := 0; j < (len(ClusterRoleRuleGroup[i].Actions)); j++ {
			if actionValidate(clusterRole.Rules, ClusterRoleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, ClusterRoleRuleGroup[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

func GetRoleRules(namespace string, name string) ([]Rule, error) {
	role, err := GetRole(namespace, name)
	if err != nil {
		return nil, err
	}

	rules := make([]Rule, 0)
	for i := 0; i < len(RoleRuleGroup); i++ {
		rule := Rule{Name: RoleRuleGroup[i].Name}
		rule.Actions = make([]Action, 0)
		for j := 0; j < len(RoleRuleGroup[i].Actions); j++ {
			if actionValidate(role.Rules, RoleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, RoleRuleGroup[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

func actionValidate(rules []v1.PolicyRule, action Action) bool {
	for _, rule := range action.Rules {
		if !ruleValidate(rules, rule) {
			return false
		}
	}
	return true
}
