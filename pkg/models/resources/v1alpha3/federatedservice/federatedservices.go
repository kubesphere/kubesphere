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

package federatedservice

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type federatedServiceGetter struct {
	informer informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) v1alpha3.Interface {
	return &federatedServiceGetter{
		informer: informer,
	}
}

func (f *federatedServiceGetter) Get(namespace, name string) (runtime.Object, error) {
	return f.informer.Types().V1beta1().FederatedServices().Lister().FederatedServices(namespace).Get(name)
}

func (f *federatedServiceGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	federatedServices, err := f.informer.Types().V1beta1().FederatedServices().Lister().FederatedServices(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, fedService := range federatedServices {
		result = append(result, fedService)
	}

	return v1alpha3.DefaultList(result, query, f.compare, f.filter), nil
}

func (f *federatedServiceGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftService, ok := left.(*v1beta1.FederatedService)
	if !ok {
		return false
	}

	rightService, ok := right.(*v1beta1.FederatedService)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftService.ObjectMeta, rightService.ObjectMeta, field)
}

func (f *federatedServiceGetter) filter(object runtime.Object, filter query.Filter) bool {
	service, ok := object.(*v1beta1.FederatedService)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(service.ObjectMeta, filter)
}
