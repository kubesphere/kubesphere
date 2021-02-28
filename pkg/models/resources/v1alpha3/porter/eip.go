/*
Copyright 2021 The KubeSphere Authors.

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

package porter

import (
	"context"
	eip "github.com/kubesphere/porter/api/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type eipGetter struct {
	c cache.Cache
}

func NewEipGetter(c cache.Cache) v1alpha3.Interface {
	return &eipGetter{c}
}

func (d *eipGetter) Get(namespace, name string) (runtime.Object, error) {
	e := eip.Eip{}
	err := d.c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (n eipGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	eips := eip.EipList{}
	err := n.c.List(context.Background(), &eips, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for i, _ := range eips.Items {
		result = append(result, &eips.Items[i])
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n eipGetter) filter(item runtime.Object, filter query.Filter) bool {
	e, ok := item.(*eip.Eip)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(e.ObjectMeta, filter)
}

func (n eipGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	l, ok := left.(*eip.Eip)
	if !ok {
		return false
	}

	r, ok := right.(*eip.Eip)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(l.ObjectMeta, r.ObjectMeta, field)
}
