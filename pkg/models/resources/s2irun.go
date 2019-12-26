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
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/kubesphere/s2ioperator/pkg/apis/devops/v1alpha1"
)

type s2iRunSearcher struct {
}

func (*s2iRunSearcher) get(namespace, name string) (interface{}, error) {
	return informers.S2iSharedInformerFactory().Devops().V1alpha1().S2iRuns().Lister().S2iRuns(namespace).Get(name)
}

// exactly Match
func (*s2iRunSearcher) match(kv map[string]string, item *v1alpha1.S2iRun) bool {
	for k, v := range kv {
		switch k {
		case Status:
			if string(item.Status.RunState) != v {
				return false
			}
		default:
			if !match(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

// Fuzzy searchInNamespace
func (*s2iRunSearcher) fuzzy(kv map[string]string, item *v1alpha1.S2iRun) bool {
	for k, v := range kv {
		if !fuzzy(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*s2iRunSearcher) compare(a, b *v1alpha1.S2iRun, orderBy string) bool {
	return compare(a.ObjectMeta, b.ObjectMeta, orderBy)
}

func (s *s2iRunSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	s2iRuns, err := informers.S2iSharedInformerFactory().Devops().V1alpha1().S2iRuns().Lister().S2iRuns(namespace).List(labels.Everything())

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
