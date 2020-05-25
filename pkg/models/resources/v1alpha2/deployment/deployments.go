/*
Copyright 2020 The KubeSphere Authors.

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

package deployment

import (
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/api/apps/v1"
)

type deploymentSearcher struct {
	informer informers.SharedInformerFactory
}

func NewDeploymentSetSearcher(informers informers.SharedInformerFactory) v1alpha2.Interface {
	return &deploymentSearcher{informer: informers}
}

func (s *deploymentSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informer.Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)
}

func deploymentStatus(item *v1.Deployment) string {
	if item.Spec.Replicas != nil {
		if item.Status.ReadyReplicas == 0 && *item.Spec.Replicas == 0 {
			return v1alpha2.StatusStopped
		} else if item.Status.ReadyReplicas == *item.Spec.Replicas {
			return v1alpha2.StatusRunning
		} else {
			return v1alpha2.StatusUpdating
		}
	}
	return v1alpha2.StatusStopped
}

func (*deploymentSearcher) match(kv map[string]string, item *v1.Deployment) bool {
	for k, v := range kv {
		switch k {
		case v1alpha2.Status:
			if deploymentStatus(item) != v {
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

func (*deploymentSearcher) fuzzy(kv map[string]string, item *v1.Deployment) bool {
	for k, v := range kv {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *deploymentSearcher) compare(a, b *v1.Deployment, orderBy string) bool {
	switch orderBy {
	case v1alpha2.UpdateTime:
		aLastUpdateTime := s.lastUpdateTime(a)
		bLastUpdateTime := s.lastUpdateTime(b)
		if aLastUpdateTime.Equal(bLastUpdateTime) {
			return strings.Compare(a.Name, b.Name) <= 0
		}
		return aLastUpdateTime.Before(bLastUpdateTime)
	default:
		return v1alpha2.ObjectMetaCompare(a.ObjectMeta, b.ObjectMeta, orderBy)
	}
}

func (s *deploymentSearcher) lastUpdateTime(deployment *v1.Deployment) time.Time {
	lastUpdateTime := deployment.CreationTimestamp.Time
	for _, condition := range deployment.Status.Conditions {
		if condition.LastUpdateTime.After(lastUpdateTime) {
			lastUpdateTime = condition.LastUpdateTime.Time
		}
	}
	return lastUpdateTime
}

func (s *deploymentSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	deployments, err := s.informer.Apps().V1().Deployments().Lister().Deployments(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.Deployment, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = deployments
	} else {
		for _, item := range deployments {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		return s.compare(result[i], result[j], orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
