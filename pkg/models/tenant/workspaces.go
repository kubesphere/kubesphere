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
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"
)

type workspaceSearcher struct {
}

// Exactly Match
func (*workspaceSearcher) match(match map[string]string, item *v1alpha1.Workspace) bool {
	for k, v := range match {
		switch k {
		case resources.Name:
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, item.Name) {
				return false
			}
		case resources.Keyword:
			if !strings.Contains(item.Name, v) && !contains(item.Labels, "", v) && !contains(item.Annotations, "", v) {
				return false
			}
		default:
			// label not exist or value not equal
			if val, ok := item.Labels[k]; !ok || val != v {
				return false
			}
		}
	}
	return true
}

func (*workspaceSearcher) fuzzy(fuzzy map[string]string, item *v1alpha1.Workspace) bool {

	for k, v := range fuzzy {
		switch k {
		case resources.Name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], v) {
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
	case resources.CreateTime:
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case resources.Name:
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

func contains(m map[string]string, key, value string) bool {
	for k, v := range m {
		if key == "" {
			if strings.Contains(k, value) || strings.Contains(v, value) {
				return true
			}
		} else if k == key && strings.Contains(v, value) {
			return true
		}
	}
	return false
}
