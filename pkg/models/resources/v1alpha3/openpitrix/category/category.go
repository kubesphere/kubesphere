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
package category

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
)

type helmCategoriesGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informers externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &helmCategoriesGetter{
		informers: informers,
	}
}

func (r *helmCategoriesGetter) Get(_, name string) (runtime.Object, error) {
	app, err := r.informers.Application().V1alpha1().HelmCategories().Lister().Get(name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return app, nil
}

func (r *helmCategoriesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var ctg []*v1alpha1.HelmCategory
	var err error

	ctg, err = r.informers.Application().V1alpha1().HelmCategories().Lister().List(query.Selector())

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var result []runtime.Object
	for i := range ctg {
		result = append(result, ctg[i])
	}

	return v1alpha3.DefaultList(result, query, r.compare, r.filter), nil
}

func (r *helmCategoriesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	ctg1, ok := left.(*v1alpha1.HelmCategory)
	if !ok {
		return false
	}

	ctg2, ok := right.(*v1alpha1.HelmCategory)
	if !ok {
		return false
	}
	switch field {
	case query.FieldName:
		return strings.Compare(ctg1.Spec.Name, ctg2.Spec.Name) > 0
	default:
		return v1alpha3.DefaultObjectMetaCompare(ctg1.ObjectMeta, ctg2.ObjectMeta, field)
	}
}

func (r *helmCategoriesGetter) filter(object runtime.Object, filter query.Filter) bool {
	application, ok := object.(*v1alpha1.HelmCategory)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(application.Spec.Name, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(application.ObjectMeta, filter)
	}
}
