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

package s2buildertemplate

import (
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"

	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
)

type s2iBuilderTemplateSearcher struct {
	informers externalversions.SharedInformerFactory
}

func NewS2iBuidlerTemplateSearcher(informers externalversions.SharedInformerFactory) v1alpha2.Interface {
	return &s2iBuilderTemplateSearcher{informers: informers}
}

func (s *s2iBuilderTemplateSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informers.Devops().V1alpha1().S2iBuilderTemplates().Lister().Get(name)
}

func (*s2iBuilderTemplateSearcher) match(match map[string]string, item *v1alpha1.S2iBuilderTemplate) bool {
	for k, v := range match {
		if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*s2iBuilderTemplateSearcher) fuzzy(fuzzy map[string]string, item *v1alpha1.S2iBuilderTemplate) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *s2iBuilderTemplateSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	builderTemplates, err := s.informers.Devops().V1alpha1().S2iBuilderTemplates().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1alpha1.S2iBuilderTemplate, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = builderTemplates
	} else {
		for _, item := range builderTemplates {
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
