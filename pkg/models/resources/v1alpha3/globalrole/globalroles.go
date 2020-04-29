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
package globalrole

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type globalrolesGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &globalrolesGetter{sharedInformers: sharedInformers}
}

func (d *globalrolesGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.Iam().V1alpha2().GlobalRoles().Lister().Get(name)
}

func (d *globalrolesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	all, err := d.sharedInformers.Iam().V1alpha2().GlobalRoles().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range all {
		result = append(result, deploy)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *globalrolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftRole, ok := left.(*iamv1alpha2.GlobalRole)
	if !ok {
		return false
	}

	rightRole, ok := right.(*iamv1alpha2.GlobalRole)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRole.ObjectMeta, rightRole.ObjectMeta, field)
}

func (d *globalrolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*iamv1alpha2.GlobalRole)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
