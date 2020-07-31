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

package federatedpersistentvolumeclaim

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	storageClassName = "storageClassName"
)

type fedPersistentVolumeClaimGetter struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedPersistentVolumeClaimGetter{informers: informer}
}

func (p *fedPersistentVolumeClaimGetter) Get(namespace, name string) (runtime.Object, error) {
	return p.informers.Types().V1beta1().FederatedPersistentVolumeClaims().Lister().FederatedPersistentVolumeClaims(namespace).Get(name)

}

func (p *fedPersistentVolumeClaimGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := p.informers.Types().V1beta1().FederatedPersistentVolumeClaims().Lister().FederatedPersistentVolumeClaims(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pvc := range all {
		result = append(result, pvc)
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *fedPersistentVolumeClaimGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshot, ok := left.(*v1beta1.FederatedPersistentVolumeClaim)
	if !ok {
		return false
	}
	rightSnapshot, ok := right.(*v1beta1.FederatedPersistentVolumeClaim)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshot.ObjectMeta, rightSnapshot.ObjectMeta, field)
}

func (p *fedPersistentVolumeClaimGetter) filter(object runtime.Object, filter query.Filter) bool {
	pvc, ok := object.(*v1beta1.FederatedPersistentVolumeClaim)
	if !ok {
		return false
	}

	switch filter.Field {
	case storageClassName:
		return pvc.Spec.Template.Spec.StorageClassName != nil && *pvc.Spec.Template.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pvc.ObjectMeta, filter)
	}
}
