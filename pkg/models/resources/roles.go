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
package resources

import (
	"sort"
	"strings"

	rbac "k8s.io/api/rbac/v1"
	lister "k8s.io/client-go/listers/rbac/v1"

	"k8s.io/apimachinery/pkg/labels"
)

type roleSearcher struct {
	roleLister lister.RoleLister
}

// exactly match
func (*roleSearcher) match(match map[string]string, item *rbac.Role) bool {
	for k, v := range match {
		switch k {
		case name:
			if item.Name != v && item.Labels[displayName] != v {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// fuzzy searchInNamespace
func (*roleSearcher) fuzzy(fuzzy map[string]string, item *rbac.Role) bool {
	for k, v := range fuzzy {
		switch k {
		case name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Labels[displayName], v) {
				return false
			}
		case label:
			if !searchFuzzy(item.Labels, "", v) {
				return false
			}
		case annotation:
			if !searchFuzzy(item.Annotations, "", v) {
				return false
			}
			return false
		case keyword:
			if !strings.Contains(item.Name, v) && !searchFuzzy(item.Labels, "", v) && !searchFuzzy(item.Annotations, "", v) {
				return false
			}
		default:
			if !searchFuzzy(item.Labels, k, v) && !searchFuzzy(item.Annotations, k, v) {
				return false
			}
		}
	}
	return true
}

func (*roleSearcher) compare(a, b *rbac.Role, orderBy string) bool {
	switch orderBy {
	case createTime:
		return a.CreationTimestamp.Time.After(b.CreationTimestamp.Time)
	case name:
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (s *roleSearcher) search(namespace string, conditions *conditions, orderBy string, reverse bool) ([]interface{}, error) {
	roles, err := s.roleLister.Roles(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*rbac.Role, 0)

	if len(conditions.match) == 0 && len(conditions.fuzzy) == 0 {
		result = roles
	} else {
		for _, item := range roles {
			if s.match(conditions.match, item) && s.fuzzy(conditions.fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		return s.compare(result[i], result[j], orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
