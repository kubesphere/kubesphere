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

package groupbinding

import (
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const User = "user"

type groupBindingGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &groupBindingGetter{sharedInformers: sharedInformers}
}

func (d *groupBindingGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.Iam().V1alpha2().GroupBindings().Lister().Get(name)
}

func (d *groupBindingGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	groupBindings, err := d.sharedInformers.Iam().V1alpha2().GroupBindings().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, groupBinding := range groupBindings {
		result = append(result, groupBinding)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *groupBindingGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftGroupBinding, ok := left.(*v1alpha2.GroupBinding)
	if !ok {
		return false
	}

	rightGroupBinding, ok := right.(*v1alpha2.GroupBinding)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftGroupBinding.ObjectMeta, rightGroupBinding.ObjectMeta, field)
}

func (d *groupBindingGetter) filter(object runtime.Object, filter query.Filter) bool {
	groupbinding, ok := object.(*v1alpha2.GroupBinding)

	if !ok {
		return false
	}

	switch filter.Field {
	case User:
		return sliceutil.HasString(groupbinding.Users, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(groupbinding.ObjectMeta, filter)
	}
}
