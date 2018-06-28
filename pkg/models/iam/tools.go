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
)

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
