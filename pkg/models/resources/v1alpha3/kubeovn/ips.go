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
	"strings"

	v1 "github.com/alauda/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/alauda/kube-ovn/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type ipsGetter struct {
	informers externalversions.SharedInformerFactory
}

func NewIps(informer externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &ipsGetter{informers: informer}
}

func (v *ipsGetter) Get(namespace, name string) (runtime.Object, error) {
	return v.informers.Kubeovn().V1().IPs().Lister().Get(name)
}

func (v *ipsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := v.informers.Kubeovn().V1().IPs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, ips := range all {
		result = append(result, ips)
	}

	return v1alpha3.DefaultList(result, query, v.compare, v.filter), nil
}

func (v *ipsGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftIp, ok := left.(*v1.IP)
	if !ok {
		return false
	}
	rightIp, ok := right.(*v1.IP)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftIp.ObjectMeta, rightIp.ObjectMeta, field)
}

func (v *ipsGetter) filter(object runtime.Object, filter query.Filter) bool {
	ip, ok := object.(*v1.IP)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldIPAddress:
		return strings.Contains(ip.Spec.IPAddress, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(ip.ObjectMeta, filter)
	}
}
