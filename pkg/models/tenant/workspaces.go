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
package tenant

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/params"
	"sort"
	"strings"
)

type workspaceSearcher struct {
}

// Exactly Match
func (*workspaceSearcher) match(match map[string]string, item *v1alpha1.Workspace) bool {
	for k, v := range match {
		switch k {
		case "name":
			if item.Name != v && item.Labels[constants.DisplayNameLabelKey] != v {
				return false
			}
		default:
			if item.Labels[k] != v {
				return false
			}
		}
	}
	return true
}

func (*workspaceSearcher) fuzzy(fuzzy map[string]string, item *v1alpha1.Workspace) bool {

	for k, v := range fuzzy {
		switch k {
		case "name":
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Labels["displayName"], v) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func (*workspaceSearcher) compare(a, b *v1alpha1.Workspace, orderBy string) bool {
	switch orderBy {
	case "createTime":
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case "name":
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (s *workspaceSearcher) search(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1alpha1.Workspace, error) {
	rules, err := iam.GetUserClusterRules(username)

	if err != nil {
		return nil, err
	}

	workspaces := make([]*v1alpha1.Workspace, 0)

	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"list"}, APIGroups: []string{"tenant.kubesphere.io"}, Resources: []string{"workspaces"}}) {
		workspaces, err = informers.KsSharedInformerFactory().Tenant().V1alpha1().Workspaces().Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
	} else {
		workspaceRoles, err := iam.GetUserWorkspaceRoleMap(username)
		if err != nil {
			return nil, err
		}
		for k := range workspaceRoles {
			workspace, err := informers.KsSharedInformerFactory().Tenant().V1alpha1().Workspaces().Lister().Get(k)
			if err != nil {
				return nil, err
			}
			workspaces = append(workspaces, workspace)
		}
	}

	result := make([]*v1alpha1.Workspace, 0)

	for _, workspace := range workspaces {
		if s.match(conditions.Match, workspace) && s.fuzzy(conditions.Fuzzy, workspace) {
			result = append(result, workspace)
		}
	}

	// order & reverse
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		return s.compare(result[i], result[j], orderBy)
	})

	return result, nil
}

func GetWorkspace(workspaceName string) (*v1alpha1.Workspace, error) {
	return informers.KsSharedInformerFactory().Tenant().V1alpha1().Workspaces().Lister().Get(workspaceName)
}
