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

package kubeovn

import (
	v1 "github.com/alauda/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/alauda/kube-ovn/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type subnetGetter struct {
	informers externalversions.SharedInformerFactory
}

func NewSubnets(informer externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &subnetGetter{informers: informer}
}

func (v *subnetGetter) Get(namespace, name string) (runtime.Object, error) {
	return v.informers.Kubeovn().V1().Subnets().Lister().Get(name)
}

func (v *subnetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := v.informers.Kubeovn().V1().Subnets().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, subnet := range all {
		if namespace != "" {
			if sliceutil.HasString(subnet.Spec.Namespaces, namespace) {
				result = append(result, subnet)
			}
		} else {
			result = append(result, subnet)
		}
	}

	return v1alpha3.DefaultList(result, query, v.compare, v.filter), nil
}

func (v *subnetGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSubnet, ok := left.(*v1.Subnet)
	if !ok {
		return false
	}
	rightSubnet, ok := right.(*v1.Subnet)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSubnet.ObjectMeta, rightSubnet.ObjectMeta, field)
}

func (v *subnetGetter) filter(object runtime.Object, filter query.Filter) bool {
	subnet, ok := object.(*v1.Subnet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(subnet.ObjectMeta, filter)
}
