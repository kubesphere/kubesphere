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
	k8sinformers "k8s.io/client-go/informers"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"
)

type NamespaceInterface interface {
	Search(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1.Namespace, error)
	CreateNamespace(workspace string, namespace *v1.Namespace, username string) (*v1.Namespace, error)
}

type namespaceSearcher struct {
	k8s       kubernetes.Interface
	informers k8sinformers.SharedInformerFactory
	am        iam.AccessManagementInterface
}

func (s *namespaceSearcher) CreateNamespace(workspace string, namespace *v1.Namespace, username string) (*v1.Namespace, error) {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}
	if username != "" {
		namespace.Annotations[constants.CreatorAnnotationKey] = username
	}

	namespace.Labels[constants.WorkspaceLabelKey] = workspace

	return s.k8s.CoreV1().Namespaces().Create(namespace)
}

func newNamespaceOperator(k8s kubernetes.Interface, informers k8sinformers.SharedInformerFactory, am iam.AccessManagementInterface) NamespaceInterface {
	return &namespaceSearcher{k8s: k8s, informers: informers, am: am}
}

func (s *namespaceSearcher) match(match map[string]string, item *v1.Namespace) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.Name:
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, item.Name) {
				return false
			}
		case v1alpha2.Keyword:
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

func (s *namespaceSearcher) fuzzy(fuzzy map[string]string, item *v1.Namespace) bool {

	for k, v := range fuzzy {
		switch k {
		case v1alpha2.Name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], v) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func (s *namespaceSearcher) compare(a, b *v1.Namespace, orderBy string) bool {
	switch orderBy {
	case "createTime":
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case "name":
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (s *namespaceSearcher) GetNamespaces(username string) ([]*v1.Namespace, error) {

	roles, err := s.am.GetRoles("", username)

	if err != nil {
		return nil, err
	}
	namespaces := make([]*v1.Namespace, 0)
	namespaceLister := s.informers.Core().V1().Namespaces().Lister()
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

func (s *namespaceSearcher) Search(username string, conditions *params.Conditions, orderBy string, reverse bool) ([]*v1.Namespace, error) {

	rules, err := s.am.GetClusterPolicyRules(username)

	if err != nil {
		return nil, err
	}

	namespaces := make([]*v1.Namespace, 0)

	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"list"}, APIGroups: []string{"tenant.kubesphere.io"}, Resources: []string{"namespaces"}}) {
		namespaces, err = s.informers.Core().V1().Namespaces().Lister().List(labels.Everything())
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
			i, j = j, i
		}
		return s.compare(result[i], result[j], orderBy)
	})

	return result, nil
}

func CreateNamespace() {

}
