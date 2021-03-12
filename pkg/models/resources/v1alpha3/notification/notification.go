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

package notification

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
)

type configGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewNotificationConfigGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &configGetter{ksInformer: informer}
}

func (g *configGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2beta1().Configs().Lister().Get(name)
}

func (g *configGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2beta1().Configs().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

type receiverGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func NewNotificationReceiverGetter(informer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &receiverGetter{ksInformer: informer}
}

func (g *receiverGetter) Get(_, name string) (runtime.Object, error) {
	return g.ksInformer.Notification().V2beta1().Receivers().Lister().Get(name)
}

func (g *receiverGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	objs, err := g.ksInformer.Notification().V2beta1().Receivers().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}
	return v1alpha3.DefaultList(result, query, compare, filter), nil
}

func compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftObj, err := meta.Accessor(left)
	if err != nil {
		return false
	}

	rightObj, err := meta.Accessor(right)
	if err != nil {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(meta.AsPartialObjectMetadata(leftObj).ObjectMeta,
		meta.AsPartialObjectMetadata(rightObj).ObjectMeta, field)
}

func filter(object runtime.Object, filter query.Filter) bool {

	accessor, err := meta.Accessor(object)
	if err != nil {
		return false
	}

	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if accessor.GetName() == name {
				return true
			}
		}
		return false
	case query.FieldName:
		return strings.Contains(accessor.GetName(), string(filter.Value))
	default:
		return true
	}
}
