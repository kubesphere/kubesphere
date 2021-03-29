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

package serviceaccount

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type serviceaccountsGetter struct {
	informer informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &serviceaccountsGetter{informer: sharedInformers}
}

func (d *serviceaccountsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.informer.Core().V1().ServiceAccounts().Lister().ServiceAccounts(namespace).Get(name)
}

func (d *serviceaccountsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	serviceaccounts, err := d.informer.Core().V1().ServiceAccounts().Lister().ServiceAccounts(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, serviceaccount := range serviceaccounts {
		result = append(result, serviceaccount)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *serviceaccountsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftCM, ok := left.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}

func (d *serviceaccountsGetter) filter(object runtime.Object, filter query.Filter) bool {
	serviceAccount, ok := object.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(serviceAccount.ObjectMeta, filter)
}
