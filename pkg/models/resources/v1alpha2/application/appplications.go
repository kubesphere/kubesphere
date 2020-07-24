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

package application

import (
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sigs.k8s.io/application/pkg/apis/app/v1beta1"
	"sigs.k8s.io/application/pkg/client/informers/externalversions"
	"sort"
)

type appSearcher struct {
	informer externalversions.SharedInformerFactory
}

func NewApplicationSearcher(informers externalversions.SharedInformerFactory) v1alpha2.Interface {
	return &appSearcher{informer: informers}
}

func (s *appSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informer.App().V1beta1().Applications().Lister().Applications(namespace).Get(name)
}

func (s *appSearcher) match(match map[string]string, item *v1beta1.Application) bool {
	for k, v := range match {
		if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *appSearcher) fuzzy(fuzzy map[string]string, item *v1beta1.Application) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *appSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	apps, err := s.informer.App().V1beta1().Applications().Lister().Applications(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1beta1.Application, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = apps
	} else {
		for _, item := range apps {
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
