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

package namespace

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type namespacesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &namespacesGetter{cache: cache}
}

func (n namespacesGetter) Get(_, name string) (runtime.Object, error) {
	namespace := &corev1.Namespace{}
	return namespace, n.cache.Get(context.Background(), types.NamespacedName{Name: name}, namespace)
}

func (n namespacesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	namespaces := &corev1.NamespaceList{}
	if err := n.cache.List(context.Background(), namespaces, client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range namespaces.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n namespacesGetter) filter(item runtime.Object, filter query.Filter) bool {
	namespace, ok := item.(*corev1.Namespace)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(string(namespace.Status.Phase), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(namespace.ObjectMeta, filter)
	}
}

func (n namespacesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNs, ok := left.(*corev1.Namespace)
	if !ok {
		return false
	}

	rightNs, ok := right.(*corev1.Namespace)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftNs.ObjectMeta, rightNs.ObjectMeta, field)
}
