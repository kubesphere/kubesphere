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

package s2irun

import (
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
)

type s2iRunSearcher struct {
	informers externalversions.SharedInformerFactory
}

func NewS2iRunSearcher(informers externalversions.SharedInformerFactory) v1alpha2.Interface {
	return &s2iRunSearcher{informers: informers}
}

func (s *s2iRunSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informers.Devops().V1alpha1().S2iRuns().Lister().S2iRuns(namespace).Get(name)
}

func (*s2iRunSearcher) match(match map[string]string, item *v1alpha1.S2iRun) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.Status:
			if string(item.Status.RunState) != v {
				return false
			}
		default:
			if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

func (*s2iRunSearcher) fuzzy(fuzzy map[string]string, item *v1alpha1.S2iRun) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *s2iRunSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	s2iRuns, err := s.informers.Devops().V1alpha1().S2iRuns().Lister().S2iRuns(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1alpha1.S2iRun, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = s2iRuns
	} else {
		for _, item := range s2iRuns {
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
