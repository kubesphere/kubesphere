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
package applicationversion

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

type applicationVersionsGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informers externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &applicationVersionsGetter{
		informers: informers,
	}
}

func (r *applicationVersionsGetter) Get(_, name string) (runtime.Object, error) {
	app, err := r.informers.Application().V1alpha1().HelmApplicationVersions().Lister().Get(name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return app, nil
}

func (r *applicationVersionsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var apps []*v1alpha1.HelmApplicationVersion
	var err error

	apps, err = r.informers.Application().V1alpha1().HelmApplicationVersions().Lister().List(query.Selector())

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var result []runtime.Object
	for i := range apps {
		result = append(result, apps[i])
	}

	return v1alpha3.DefaultList(result, query, r.compare, r.filter), nil
}

func (r *applicationVersionsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftAppVer, ok := left.(*v1alpha1.HelmApplicationVersion)
	if !ok {
		return false
	}

	rightAppVer, ok := right.(*v1alpha1.HelmApplicationVersion)
	if !ok {
		return false
	}
	switch field {
	case query.FieldName:
		return strings.Compare(leftAppVer.Spec.Name, rightAppVer.Spec.Name) > 0
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftAppVer.ObjectMeta, rightAppVer.ObjectMeta, field)
	}
}

func (r *applicationVersionsGetter) filter(object runtime.Object, filter query.Filter) bool {
	appVer, ok := object.(*v1alpha1.HelmApplicationVersion)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(appVer.Spec.Name, string(filter.Value))
	case query.FieldStatus:
		return strings.Contains(appVer.Status.State, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(appVer.ObjectMeta, filter)
	}
}
