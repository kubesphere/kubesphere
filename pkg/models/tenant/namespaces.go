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
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"
)

type namespaceSearcher struct {
}

// Exactly Match
func (*namespaceSearcher) match(match map[string]string, item *v1.Namespace) bool {
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

func (*namespaceSearcher) fuzzy(fuzzy map[string]string, item *v1.Namespace) bool {

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

func (*namespaceSearcher) compare(a, b *v1.Namespace, orderBy string) bool {
	switch orderBy {
	case "createTime":
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case "name":
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (*namespaceSearcher) GetNamespaces(username string) ([]*v1.Namespace, error) {

	roles, err := iam.GetUserRoles("", username)

	if err != nil {
		return nil, err
	}
	namespaces := make([]*v1.Namespace, 0)
	namespaceLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	for _, role := range roles {
		namespace, err := namespaceLister.Get(role.Namespace)
		if err != nil {
			klog.Errorf("get namespace failed: %+v", err)
			return nil, err
		}
		if !containsNamespace(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	return namespaces, nil
}

func containsNamespace(namespaces []*v1.Namespace, namespace *v1.Namespace) bool {
	for _, item := range namespaces {
		if item.Name == namespace.Name {
			return true
		}
	}
	return false
}

func (s *namespaceSearcher) search(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1.Namespace, error) {

	rules, err := iam.GetUserClusterRules(username)

	if err != nil {
		return nil, err
	}

	namespaces := make([]*v1.Namespace, 0)

	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"list"}, APIGroups: []string{"tenant.kubesphere.io"}, Resources: []string{"namespaces"}}) {
		namespaces, err = informers.SharedInformerFactory().Core().V1().Namespaces().Lister().List(labels.Everything())
	} else {
		namespaces, err = s.GetNamespaces(username)
	}

	if err != nil {
		return nil, err
	}

	result := make([]*v1.Namespace, 0)

	for _, namespace := range namespaces {
		if s.match(conditions.Match, namespace) && s.fuzzy(conditions.Fuzzy, namespace) {
			result = append(result, namespace)
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
