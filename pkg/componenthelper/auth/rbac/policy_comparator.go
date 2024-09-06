/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// This file is copied from K8s library .
// https://k8s.io/component-helpers/auth/rbac/validation/policy_comparator.go

package rbac

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
)

func Covers(ownerRules, servantRules []rbacv1.PolicyRule) (bool, []rbacv1.PolicyRule) {
	// 1.  Break every servantRule into individual rule tuples: group, verb, resource, resourceName
	// 2.  Compare the mini-rules against each owner rule.  Because the breakdown is down to the most atomic level, we're guaranteed that each mini-servant rule will be either fully covered or not covered by a single owner rule
	// 3.  Any left over mini-rules means that we are not covered and we have a nice list of them.
	// TODO: it might be nice to collapse the list down into something more human readable

	var rules []rbacv1.PolicyRule
	for _, servantRule := range servantRules {
		rules = append(rules, BreakdownRule(servantRule)...)
	}

	var uncoveredRules []rbacv1.PolicyRule
	for _, rule := range rules {
		covered := false
		for _, ownerRule := range ownerRules {
			if ruleCovers(ownerRule, rule) {
				covered = true
				break
			}
		}

		if !covered {
			uncoveredRules = append(uncoveredRules, rule)
		}
	}

	return len(uncoveredRules) == 0, uncoveredRules
}

// BreakdownRule takes a rule and builds an equivalent list of rules that each have at most one verb, one
// resource, and one resource name
func BreakdownRule(rule rbacv1.PolicyRule) []rbacv1.PolicyRule {
	var rules []rbacv1.PolicyRule
	for _, group := range rule.APIGroups {
		for _, resource := range rule.Resources {
			for _, verb := range rule.Verbs {
				if len(rule.ResourceNames) > 0 {
					for _, resourceName := range rule.ResourceNames {
						rules = append(rules, rbacv1.PolicyRule{APIGroups: []string{group}, Resources: []string{resource}, Verbs: []string{verb}, ResourceNames: []string{resourceName}})
					}

				} else {
					rules = append(rules, rbacv1.PolicyRule{APIGroups: []string{group}, Resources: []string{resource}, Verbs: []string{verb}})
				}

			}
		}
	}

	// Non-resource URLs are unique because they only combine with verbs.
	for _, nonResourceURL := range rule.NonResourceURLs {
		for _, verb := range rule.Verbs {
			rules = append(rules, rbacv1.PolicyRule{NonResourceURLs: []string{nonResourceURL}, Verbs: []string{verb}})
		}
	}

	return rules
}

func has(set []string, ele string) bool {
	for _, s := range set {
		if s == ele {
			return true
		}
	}
	return false
}

func hasAll(set, contains []string) bool {
	owning := make(map[string]struct{}, len(set))
	for _, ele := range set {
		owning[ele] = struct{}{}
	}
	for _, ele := range contains {
		if _, ok := owning[ele]; !ok {
			return false
		}
	}
	return true
}

func resourceCoversAll(setResources, coversResources []string) bool {
	// if we have a star or an exact match on all resources, then we match
	if has(setResources, rbacv1.ResourceAll) || hasAll(setResources, coversResources) {
		return true
	}

	for _, path := range coversResources {
		// if we have an exact match, then we match.
		if has(setResources, path) {
			continue
		}
		// if we're not a subresource, then we definitely don't match.  fail.
		if !strings.Contains(path, "/") {
			return false
		}
		tokens := strings.SplitN(path, "/", 2)
		resourceToCheck := "*/" + tokens[1]
		if !has(setResources, resourceToCheck) {
			return false
		}
	}

	return true
}

func nonResourceURLsCoversAll(set, covers []string) bool {
	for _, path := range covers {
		covered := false
		for _, owner := range set {
			if nonResourceURLCovers(owner, path) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

func nonResourceURLCovers(ownerPath, subPath string) bool {
	if ownerPath == subPath {
		return true
	}
	return strings.HasSuffix(ownerPath, "*") && strings.HasPrefix(subPath, strings.TrimRight(ownerPath, "*"))
}

// ruleCovers determines whether the ownerRule (which may have multiple verbs, resources, and resourceNames) covers
// the subrule (which may only contain at most one verb, resource, and resourceName)
func ruleCovers(ownerRule, subRule rbacv1.PolicyRule) bool {
	verbMatches := has(ownerRule.Verbs, rbacv1.VerbAll) || hasAll(ownerRule.Verbs, subRule.Verbs)
	groupMatches := has(ownerRule.APIGroups, rbacv1.APIGroupAll) || hasAll(ownerRule.APIGroups, subRule.APIGroups)
	resourceMatches := resourceCoversAll(ownerRule.Resources, subRule.Resources)
	nonResourceURLMatches := nonResourceURLsCoversAll(ownerRule.NonResourceURLs, subRule.NonResourceURLs)

	resourceNameMatches := false

	if len(subRule.ResourceNames) == 0 {
		resourceNameMatches = len(ownerRule.ResourceNames) == 0
	} else {
		resourceNameMatches = len(ownerRule.ResourceNames) == 0 || hasAll(ownerRule.ResourceNames, subRule.ResourceNames)
	}

	return verbMatches && groupMatches && resourceMatches && resourceNameMatches && nonResourceURLMatches
}
