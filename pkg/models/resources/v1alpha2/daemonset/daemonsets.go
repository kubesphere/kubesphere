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

package daemonset

import (
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
)

type daemonSetSearcher struct {
	informer informers.SharedInformerFactory
}

func NewDaemonSetSearcher(informers informers.SharedInformerFactory) v1alpha2.Interface {
	return &daemonSetSearcher{informer: informers}
}

func (c *daemonSetSearcher) Get(namespace, name string) (interface{}, error) {
	return c.informer.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)
}

func daemonSetStatus(item *v1.DaemonSet) string {
	if item.Status.DesiredNumberScheduled == 0 && item.Status.NumberReady == 0 {
		return v1alpha2.StatusStopped
	} else if item.Status.DesiredNumberScheduled == item.Status.NumberReady {
		return v1alpha2.StatusRunning
	} else {
		return v1alpha2.StatusUpdating
	}
}

func (*daemonSetSearcher) match(match map[string]string, item *v1.DaemonSet) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.Status:
			if daemonSetStatus(item) != v {
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

func (*daemonSetSearcher) fuzzy(kv map[string]string, item *v1.DaemonSet) bool {
	for k, v := range kv {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (c *daemonSetSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	daemonSets, err := c.informer.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.DaemonSet, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = daemonSets
	} else {
		for _, item := range daemonSets {
			if c.match(conditions.Match, item) && c.fuzzy(conditions.Fuzzy, item) {
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
