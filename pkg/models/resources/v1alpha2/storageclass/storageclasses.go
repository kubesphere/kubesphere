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

package storageclass

import (
	snapshotinformer "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
	"strconv"
)

type storageClassesSearcher struct {
	informers         informers.SharedInformerFactory
	snapshotInformers snapshotinformer.SharedInformerFactory
}

func NewStorageClassesSearcher(informers informers.SharedInformerFactory, snapshotInformer snapshotinformer.SharedInformerFactory) v1alpha2.Interface {
	return &storageClassesSearcher{
		informers:         informers,
		snapshotInformers: snapshotInformer,
	}
}

func (s *storageClassesSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informers.Storage().V1().StorageClasses().Lister().Get(name)
}

func (*storageClassesSearcher) match(match map[string]string, item *v1.StorageClass) bool {
	for k, v := range match {
		if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*storageClassesSearcher) fuzzy(fuzzy map[string]string, item *v1.StorageClass) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *storageClassesSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	storageClasses, err := s.informers.Storage().V1().StorageClasses().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.StorageClass, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = storageClasses
	} else {
		for _, item := range storageClasses {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		return v1alpha2.ObjectMetaCompare(result[i].ObjectMeta, result[j].ObjectMeta, orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		count := s.countPersistentVolumeClaims(i.Name)
		if i.Annotations == nil {
			i.Annotations = make(map[string]string)
		}
		i.Annotations["kubesphere.io/pvc-count"] = strconv.Itoa(count)
		r = append(r, i)
	}
	return r, nil
}

func (s *storageClassesSearcher) countPersistentVolumeClaims(name string) int {
	pvcs, err := s.informers.Core().V1().PersistentVolumeClaims().Lister().List(labels.Everything())
	if err != nil {
		return 0
	}
	var count int

	for _, pvc := range pvcs {
		if *pvc.Spec.StorageClassName == name || (pvc.Annotations != nil && pvc.Annotations[corev1.BetaStorageClassAnnotation] == name) {
			count++
		}
	}

	return count
}
