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

package federatedapplication

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type fedApplicationsGetter struct {
	informer informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedApplicationsGetter{informer: sharedInformers}
}

func (d *fedApplicationsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.informer.Types().V1beta1().FederatedApplications().Lister().FederatedApplications(namespace).Get(name)
}

func (d *fedApplicationsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	applications, err := d.informer.Types().V1beta1().FederatedApplications().Lister().FederatedApplications(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, app := range applications {
		result = append(result, app)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *fedApplicationsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftApplication, ok := left.(*v1beta1.FederatedApplication)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*v1beta1.FederatedApplication)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftApplication) > (lastUpdateTime(rightApplication))
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftApplication.ObjectMeta, rightApplication.ObjectMeta, field)
	}
}

func (d *fedApplicationsGetter) filter(object runtime.Object, filter query.Filter) bool {
	application, ok := object.(*v1beta1.FederatedApplication)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(application.ObjectMeta, filter)
}

func lastUpdateTime(application *v1beta1.FederatedApplication) string {
	lut := application.CreationTimestamp.Time.String()
	for _, condition := range application.Status.Conditions {
		if condition.LastUpdateTime > lut {
			lut = condition.LastUpdateTime
		}
	}
	return lut
}
