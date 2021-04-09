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

package customresourcedefinition

import (
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type crdGetter struct {
	informers apiextensionsinformers.SharedInformerFactory
}

func New(informers apiextensionsinformers.SharedInformerFactory) v1alpha3.Interface {
	return &crdGetter{
		informers: informers,
	}
}

func (c crdGetter) Get(_, name string) (runtime.Object, error) {
	return c.informers.Apiextensions().V1().CustomResourceDefinitions().Lister().Get(name)
}

func (c crdGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	crds, err := c.informers.Apiextensions().V1().CustomResourceDefinitions().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, crd := range crds {
		result = append(result, crd)
	}

	return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
}

func (c crdGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftCRD, ok := left.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	rightCRD, ok := right.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCRD.ObjectMeta, rightCRD.ObjectMeta, field)
}

func (c crdGetter) filter(object runtime.Object, filter query.Filter) bool {
	crd, ok := object.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(crd.Name, string(filter.Value)) || strings.Contains(crd.Spec.Names.Kind, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(crd.ObjectMeta, filter)
	}
}
