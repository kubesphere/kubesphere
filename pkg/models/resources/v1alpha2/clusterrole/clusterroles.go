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

package clusterrole

import (
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"sort"
)

type clusterRoleSearcher struct {
	informer informers.SharedInformerFactory
}

func NewClusterRoleSearcher(informer informers.SharedInformerFactory) v1alpha2.Interface {
	return &clusterRoleSearcher{informer: informer}
}

func (s *clusterRoleSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informer.Rbac().V1().ClusterRoles().Lister().Get(name)
}

func (*clusterRoleSearcher) match(match map[string]string, item *rbac.ClusterRole) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.OwnerKind:
			fallthrough
		case v1alpha2.OwnerName:
			kind := match[v1alpha2.OwnerKind]
			name := match[v1alpha2.OwnerName]
			if !k8sutil.IsControlledBy(item.OwnerReferences, kind, name) {
				return false
			}
		case v1alpha2.UserFacing:
			if v == "true" {
				if !isUserFacingClusterRole(item) {
					return false
				}
			}
		default:
			if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

func (s *clusterRoleSearcher) fuzzy(fuzzy map[string]string, item *rbac.ClusterRole) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *clusterRoleSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	clusterRoles, err := s.informer.Rbac().V1().ClusterRoles().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*rbac.ClusterRole, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = clusterRoles
	} else {
		for _, item := range clusterRoles {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		return v1alpha2.ObjectMetaCompare(result[i].ObjectMeta, result[j].ObjectMeta, orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}

// cluster role created by user from kubesphere dashboard
func isUserFacingClusterRole(role *rbac.ClusterRole) bool {
	if role.Annotations[constants.CreatorAnnotationKey] != "" && role.Labels[constants.WorkspaceLabelKey] == "" {
		return true
	}
	return false
}
