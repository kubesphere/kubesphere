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
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type serviceAccountsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &serviceAccountsGetter{cache: cache}
}

func (d *serviceAccountsGetter) Get(namespace, name string) (runtime.Object, error) {
	serviceAccount := &corev1.ServiceAccount{}
	return serviceAccount, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, serviceAccount)
}

func (d *serviceAccountsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	serviceAccounts := &corev1.ServiceAccountList{}
	if err := d.cache.List(context.Background(), serviceAccounts, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range serviceAccounts.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *serviceAccountsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftSA, ok := left.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	rightSA, ok := right.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftSA.ObjectMeta, rightSA.ObjectMeta, field)
}

func (d *serviceAccountsGetter) filter(object runtime.Object, filter query.Filter) bool {
	serviceAccount, ok := object.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(serviceAccount.ObjectMeta, filter)
}
