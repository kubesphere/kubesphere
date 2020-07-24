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

package application

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	appv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
	"sigs.k8s.io/application/pkg/client/informers/externalversions"
	"time"
)

type applicationsGetter struct {
	informer externalversions.SharedInformerFactory
}

func New(sharedInformers externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &applicationsGetter{informer: sharedInformers}
}

func (d *applicationsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.informer.App().V1beta1().Applications().Lister().Applications(namespace).Get(name)
}

func (d *applicationsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	applications, err := d.informer.App().V1beta1().Applications().Lister().Applications(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, app := range applications {
		result = append(result, app)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *applicationsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftApplication, ok := left.(*appv1beta1.Application)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*appv1beta1.Application)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftApplication).After(lastUpdateTime(rightApplication))
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftApplication.ObjectMeta, rightApplication.ObjectMeta, field)
	}
}

func (d *applicationsGetter) filter(object runtime.Object, filter query.Filter) bool {
	application, ok := object.(*appv1beta1.Application)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(application.ObjectMeta, filter)
}

func lastUpdateTime(application *appv1beta1.Application) time.Time {
	lut := application.CreationTimestamp.Time
	for _, condition := range application.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
