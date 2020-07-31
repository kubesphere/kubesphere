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

package federatedstatefulset

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type fedStatefulSetGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedStatefulSetGetter{sharedInformers: sharedInformers}
}

func (d *fedStatefulSetGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Types().V1beta1().FederatedStatefulSets().Lister().FederatedStatefulSets(namespace).Get(name)
}

func (d *fedStatefulSetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	// first retrieves all statefulSets within given namespace
	statefulSets, err := d.sharedInformers.Types().V1beta1().FederatedStatefulSets().Lister().FederatedStatefulSets(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, statefulSet := range statefulSets {
		result = append(result, statefulSet)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *fedStatefulSetGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftStatefulSet, ok := left.(*v1beta1.FederatedStatefulSet)
	if !ok {
		return false
	}

	rightStatefulSet, ok := right.(*v1beta1.FederatedStatefulSet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftStatefulSet.ObjectMeta, rightStatefulSet.ObjectMeta, field)
}

func (d *fedStatefulSetGetter) filter(object runtime.Object, filter query.Filter) bool {
	statefulSet, ok := object.(*v1beta1.FederatedStatefulSet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(statefulSet.ObjectMeta, filter)
}
