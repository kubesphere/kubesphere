/*
Copyright 2020 KubeSphere Authors

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

package group

import (
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type groupGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &groupGetter{sharedInformers: sharedInformers}
}

func (d *groupGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.Iam().V1alpha2().Groups().Lister().Get(name)
}

func (d *groupGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	groups, err := d.sharedInformers.Iam().V1alpha2().Groups().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, group := range groups {
		result = append(result, group)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *groupGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftGroup, ok := left.(*v1alpha2.Group)
	if !ok {
		return false
	}

	rightGroup, ok := right.(*v1alpha2.Group)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftGroup.ObjectMeta, rightGroup.ObjectMeta, field)
}

func (d *groupGetter) filter(object runtime.Object, filter query.Filter) bool {
	group, ok := object.(*v1alpha2.Group)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(group.ObjectMeta, filter)
}
