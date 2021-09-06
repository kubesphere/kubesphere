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

package persistentvolume

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	storageClassName = "storageClassName"
)

type persistentVolumeGetter struct {
	informers informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) v1alpha3.Interface {
	return &persistentVolumeGetter{informers: informer}
}

func (p *persistentVolumeGetter) Get(namespace, name string) (runtime.Object, error) {
	pv, err := p.informers.Core().V1().PersistentVolumes().Lister().Get(name)
	return pv, err
}

func (p *persistentVolumeGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := p.informers.Core().V1().PersistentVolumes().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, pv := range all {
		result = append(result, pv)
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *persistentVolumeGetter) compare(obj1, obj2 runtime.Object, field query.Field) bool {
	pv1, ok := obj1.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	pv2, ok := obj2.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(pv1.ObjectMeta, pv2.ObjectMeta, field)
}

func (p *persistentVolumeGetter) filter(object runtime.Object, filter query.Filter) bool {
	pv, ok := object.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(pv.Status.Phase), string(filter.Value))
	case storageClassName:
		return pv.Spec.StorageClassName != "" && pv.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pv.ObjectMeta, filter)
	}
}
