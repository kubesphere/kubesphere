/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package iam

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"strings"
)

func RulesMatchesRequired(rules []rbacv1.PolicyRule, required rbacv1.PolicyRule) bool {
	for _, rule := range rules {
		if ruleMatchesRequired(rule, required) {
			return true
		}
	}
	return false
}

func rulesMatchesAction(rules []rbacv1.PolicyRule, action policy.Action) bool {

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

func ContainsUser(subjects interface{}, username string) bool {
	switch subjects.(type) {
	case []*rbacv1.Subject:
		for _, subject := range subjects.([]*rbacv1.Subject) {
			if subject.Kind == rbacv1.UserKind && subject.Name == username {
				return true
			}
		}
	case []rbacv1.Subject:
		for _, subject := range subjects.([]rbacv1.Subject) {
			if subject.Kind == rbacv1.UserKind && subject.Name == username {
				return true
			}
		}
	case []User:
		for _, u := range subjects.([]User) {
			if u.Username == username {
				return true
			}
		}

	case []*User:
		for _, u := range subjects.([]*User) {
			if u.Username == username {
				return true
			}
		}
	}
	return false
}
