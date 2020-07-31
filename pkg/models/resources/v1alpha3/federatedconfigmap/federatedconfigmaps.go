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

package federatedconfigmap

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type fedConfigMapsGetter struct {
	informer informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedConfigMapsGetter{informer: sharedInformers}
}

func (d *fedConfigMapsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.informer.Types().V1beta1().FederatedConfigMaps().Lister().FederatedConfigMaps(namespace).Get(name)
}

func (d *fedConfigMapsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	configmaps, err := d.informer.Types().V1beta1().FederatedConfigMaps().Lister().FederatedConfigMaps(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, configmap := range configmaps {
		result = append(result, configmap)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *fedConfigMapsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftCM, ok := left.(*v1beta1.FederatedConfigMap)
	if !ok {
		return false
	}

	rightCM, ok := right.(*v1beta1.FederatedConfigMap)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}

func (d *fedConfigMapsGetter) filter(object runtime.Object, filter query.Filter) bool {
	configMap, ok := object.(*v1beta1.FederatedConfigMap)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(configMap.ObjectMeta, filter)
}
