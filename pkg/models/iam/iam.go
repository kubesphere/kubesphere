package iam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	v12 "k8s.io/client-go/listers/rbac/v1"

	"kubesphere.io/kubesphere/pkg/informers"

	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/util/slice"

	"kubesphere.io/kubesphere/pkg/constants"
	ksErr "kubesphere.io/kubesphere/pkg/errors"
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

func GetRole(namespace string, name string) (*v1.Role, error) {
	role, err := roleLister.Roles(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	return role.DeepCopy(), nil
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

func GetClusterRole(name string) (*v1.ClusterRole, error) {

	role, err := clusterRoleLister.Get(name)

	if err != nil {
		return nil, err
	}
	return role.DeepCopy(), nil
}

func GetRoles(namespace string, username string) ([]v1.Role, error) {
	roleBindings, err := roleBindingLister.RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	roles := make([]v1.Role, 0)

	for _, roleBinding := range roleBindings {

		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {
					clusterRole, err := clusterRoleLister.Get(roleBinding.RoleRef.Name)
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
						role, err := roleLister.Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name)
						if err == nil {
							roles = append(roles, *role)
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

// Get cluster roles by username
func GetClusterRoles(username string) ([]v1.ClusterRole, error) {
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	roles := make([]v1.ClusterRole, 0)

	for _, roleBinding := range clusterRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {
					role, err := clusterRoleLister.Get(roleBinding.RoleRef.Name)
					if err == nil {
						role = role.DeepCopy()
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
						glog.Warning(err)
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

func RulesMatchesRequired(rules []v1.PolicyRule, required v1.PolicyRule) bool {
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
