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

package federatednamespace

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	typesv1beta1 "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type federatedNamespacesGetter struct {
	informers informers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory) v1alpha3.Interface {
	return &federatedNamespacesGetter{informers: informers}
}

func (n federatedNamespacesGetter) Get(namespace, name string) (runtime.Object, error) {
	return n.informers.Types().V1beta1().FederatedNamespaces().Lister().FederatedNamespaces(namespace).Get(name)
}

func (n federatedNamespacesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	ns, err := n.informers.Types().V1beta1().FederatedNamespaces().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range ns {
		result = append(result, item)
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n federatedNamespacesGetter) filter(item runtime.Object, filter query.Filter) bool {
	namespace, ok := item.(*typesv1beta1.FederatedNamespace)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaFilter(namespace.ObjectMeta, filter)
}

func (n federatedNamespacesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNs, ok := left.(*typesv1beta1.FederatedNamespace)
	if !ok {
		return false
	}

	rightNs, ok := right.(*typesv1beta1.FederatedNamespace)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftNs.ObjectMeta, rightNs.ObjectMeta, field)
}
