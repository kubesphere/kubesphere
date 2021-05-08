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

package ippool

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8sinformers "k8s.io/client-go/informers"

	networkv1alpha1 "kubesphere.io/api/network/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type ippoolGetter struct {
	informers    informers.SharedInformerFactory
	k8sInformers k8sinformers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory, k8sInformers k8sinformers.SharedInformerFactory) v1alpha3.Interface {
	return &ippoolGetter{
		informers:    informers,
		k8sInformers: k8sInformers,
	}
}

func (n ippoolGetter) Get(namespace, name string) (runtime.Object, error) {
	return n.informers.Network().V1alpha1().IPPools().Lister().Get(name)
}

func (n ippoolGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	var result []runtime.Object

	if namespace != "" {
		workspace := ""
		ns, err := n.k8sInformers.Core().V1().Namespaces().Lister().Get(namespace)
		if err != nil {
			return nil, err
		}
		if ns.Labels != nil {
			workspace = ns.Labels[constants.WorkspaceLabelKey]
		}
		ps, err := n.informers.Network().V1alpha1().IPPools().Lister().List(labels.SelectorFromSet(
			map[string]string{
				networkv1alpha1.IPPoolDefaultLabel: "",
			}))
		if err != nil {
			return nil, err
		}
		for _, p := range ps {
			result = append(result, p)
		}
		if workspace != "" {
			query.LabelSelector = labels.SelectorFromSet(
				map[string]string{
					constants.WorkspaceLabelKey: workspace,
				}).String()
			ps, err := n.informers.Network().V1alpha1().IPPools().Lister().List(query.Selector())
			if err != nil {
				return nil, err
			}
			for _, p := range ps {
				result = append(result, p)
			}
		}
	} else {
		ps, err := n.informers.Network().V1alpha1().IPPools().Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
		for _, p := range ps {
			result = append(result, p)
		}
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n ippoolGetter) filter(item runtime.Object, filter query.Filter) bool {
	p, ok := item.(*networkv1alpha1.IPPool)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(p.ObjectMeta, filter)
}

func (n ippoolGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftP, ok := left.(*networkv1alpha1.IPPool)
	if !ok {
		return false
	}

	rightP, ok := right.(*networkv1alpha1.IPPool)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftP.ObjectMeta, rightP.ObjectMeta, field)
}
