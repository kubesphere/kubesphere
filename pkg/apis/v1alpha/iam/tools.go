/*
 Copyright 2018 The KubeSphere Authors.

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
	"k8s.io/api/rbac/v1"
	"k8s.io/kubernetes/pkg/util/slice"
	"kubesphere.io/kubesphere/pkg/models"
)

func getUserRules(username string) (map[string][]rule, error) {

	items := make(map[string][]rule, 0)
	roles, err := models.GetRoles(username)

	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0)

	for i := 0; i < len(roles); i++ {
		if !slice.ContainsString(namespaces, roles[i].Namespace, nil) {
			namespaces = append(namespaces, roles[i].Namespace)
		}
	}

	for _, namespace := range namespaces {
		rules := getMergeRules(namespace, roles)
		if len(rules) > 0 {
			items[namespace] = rules
		}
	}

	return items, nil
}

func getMergeRules(namespace string, roles []v1.Role) []rule {
	rules := make([]rule, 0)

	for i := 0; i < (len(roleRuleGroup)); i++ {
		rule := rule{Name: roleRuleGroup[i].Name}
		rule.Actions = make([]action, 0)
		for j := 0; j < (len(roleRuleGroup[i].Actions)); j++ {
			permit := false
			for _, role := range roles {
				if role.Namespace == namespace && actionValidate(role.Rules, roleRuleGroup[i].Actions[j]) {
					permit = true
					break
				}
			}
			if permit {
				rule.Actions = append(rule.Actions, roleRuleGroup[i].Actions[j])
			}
		}

		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules
}

func getUserClusterRules(username string) ([]rule, error) {

	rules := make([]rule, 0)

	roles, err := models.GetClusterRoles(username)

	if err != nil {
		return nil, err
	}

	for i := 0; i < (len(clusterRoleRuleGroup)); i++ {
		rule := rule{Name: clusterRoleRuleGroup[i].Name}
		rule.Actions = make([]action, 0)
		for j := 0; j < (len(clusterRoleRuleGroup[i].Actions)); j++ {
			actionPermit := false
			for _, role := range roles {
				if actionValidate(role.Rules, clusterRoleRuleGroup[i].Actions[j]) {
					actionPermit = true
					break
				}
			}
			if actionPermit {
				rule.Actions = append(rule.Actions, clusterRoleRuleGroup[i].Actions[j])
			}
		}

		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

func getClusterRoleRules(name string) ([]rule, error) {

	clusterRole, err := models.GetClusterRole(name)

	if err != nil {
		return nil, err
	}

	rules := make([]rule, 0)

	for i := 0; i < len(clusterRoleRuleGroup); i ++ {
		rule := rule{Name: clusterRoleRuleGroup[i].Name}
		rule.Actions = make([]action, 0)
		for j := 0; j < (len(clusterRoleRuleGroup[i].Actions)); j++ {
			if actionValidate(clusterRole.Rules, clusterRoleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, clusterRoleRuleGroup[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

func getRoleRules(namespace string, name string) ([]rule, error) {
	role, err := models.GetRole(namespace, name)
	if err != nil {
		return nil, err
	}
	rules := make([]rule, 0)
	for i := 0; i < len(roleRuleGroup); i ++ {
		rule := rule{Name: roleRuleGroup[i].Name}
		rule.Actions = make([]action, 0)
		for j := 0; j < len(roleRuleGroup[i].Actions); j++ {
			if actionValidate(role.Rules, roleRuleGroup[i].Actions[j]) {
				rule.Actions = append(rule.Actions, roleRuleGroup[i].Actions[j])
			}
		}
		if len(rule.Actions) > 0 {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

func actionValidate(rules []v1.PolicyRule, action action) bool {
	for _, rule := range action.Rules {
		if !ruleValidate(rules, rule) {
			return false
		}
	}
	return true
}

func ruleValidate(rules []v1.PolicyRule, rule v1.PolicyRule) bool {

	for _, apiGroup := range rule.APIGroups {
		if len(rule.NonResourceURLs) == 0 {
			for _, resource := range rule.Resources {

				//if len(rule.ResourceNames) == 0 {

				for _, verb := range rule.Verbs {
					if !verbValidate(rules, apiGroup, "", resource, "", verb) {
						return false
					}
				}

				//} else {
				//	for _, resourceName := range rule.ResourceNames {
				//		for _, verb := range rule.Verbs {
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
		if slice.ContainsString(rule.APIGroups, apiGroup, nil) || slice.ContainsString(rule.APIGroups, v1.APIGroupAll, nil) {
			if slice.ContainsString(rule.Verbs, verb, nil) || slice.ContainsString(rule.Verbs, v1.VerbAll, nil) {
				if nonResourceURL == "" {
					if slice.ContainsString(rule.Resources, resource, nil) || slice.ContainsString(rule.Resources, v1.ResourceAll, nil) {
						if resourceName == "" {
							return true
						} else if slice.ContainsString(rule.ResourceNames, resourceName, nil) || slice.ContainsString(rule.Resources, v1.ResourceAll, nil) {
							return true
						}
					}
				} else if slice.ContainsString(rule.NonResourceURLs, nonResourceURL, nil) || slice.ContainsString(rule.NonResourceURLs, v1.NonResourceAll, nil) {
					return true
				}
			}
		}
	}
	return false
}
