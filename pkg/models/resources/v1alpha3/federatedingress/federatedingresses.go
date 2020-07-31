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
package federatedingress

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type fedIngressGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedIngressGetter{sharedInformers: sharedInformers}
}

func (g *fedIngressGetter) Get(namespace, name string) (runtime.Object, error) {
	return g.sharedInformers.Types().V1beta1().FederatedIngresses().Lister().FederatedIngresses(namespace).Get(name)
}

func (g *fedIngressGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	// first retrieves all deployments within given namespace
	ingresses, err := g.sharedInformers.Types().V1beta1().FederatedIngresses().Lister().FederatedIngresses(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, ingress := range ingresses {
		result = append(result, ingress)
	}

	return v1alpha3.DefaultList(result, query, g.compare, g.filter), nil
}

func (g *fedIngressGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftIngress, ok := left.(*v1beta1.FederatedIngress)
	if !ok {
		return false
	}

	rightIngress, ok := right.(*v1beta1.FederatedIngress)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftIngress.ObjectMeta, rightIngress.ObjectMeta, field)
}

func (g *fedIngressGetter) filter(object runtime.Object, filter query.Filter) bool {
	deployment, ok := object.(*v1beta1.FederatedIngress)
	if !ok {
		return false
	}

	switch filter.Field {
	default:
		return v1alpha3.DefaultObjectMetaFilter(deployment.ObjectMeta, filter)
	}
}
