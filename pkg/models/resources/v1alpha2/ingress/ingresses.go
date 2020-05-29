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

package ingress

import (
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"

	"k8s.io/apimachinery/pkg/labels"
)

type ingressSearcher struct {
	informers informers.SharedInformerFactory
}

func NewIngressSearcher(informers informers.SharedInformerFactory) v1alpha2.Interface {
	return &ingressSearcher{informers: informers}
}

func (s *ingressSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informers.Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).Get(name)
}

func (*ingressSearcher) match(match map[string]string, item *v1beta1.Ingress) bool {
	for k, v := range match {
		if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*ingressSearcher) fuzzy(fuzzy map[string]string, item *v1beta1.Ingress) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *ingressSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	ingresses, err := s.informers.Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1beta1.Ingress, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = ingresses
	} else {
		for _, item := range ingresses {
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
