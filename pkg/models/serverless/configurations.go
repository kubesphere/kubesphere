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

package serverless

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type configSearcher struct {
}

func (s *configSearcher) match(match map[string]string, item *v1beta1.Configuration) bool {
	return true
}

func (s *configSearcher) fuzzy(match map[string]string, item *v1beta1.Configuration) bool {
	return true
}

func (s *configSearcher) compare(a, b *v1beta1.Configuration, orderBy string) bool {
	return true
}

func (s *configSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	result, err := informers.ServerlessInformerFactory().Serving().V1beta1().Configurations().Lister().Configurations(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	/*
	result := make([]*v1beta1.Service, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = services
	} else {
		for _, item := range services {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
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
	*/

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}

var (
	configurations = configSearcher{}
)

func ListConfigurations(namespace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)

	result, err := configurations.search(namespace, conditions, orderBy, reverse)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}

	result = result[offset : offset+limit]
	for _, item := range result {
		items = append(items, item)
	}
	return &models.PageableResponse{Items: items, TotalCount: len(items)}, nil
}
