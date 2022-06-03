/*
Copyright 2021 The KubeSphere Authors.

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

package volumesnapshotclass

import (
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	deletionPolicy = "deletionPolicy"
	driver         = "driver"
)

type volumeSnapshotClassGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informer externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &volumeSnapshotClassGetter{informers: informer}
}

func (v *volumeSnapshotClassGetter) Get(namespace, name string) (runtime.Object, error) {
	return v.informers.Snapshot().V1().VolumeSnapshotClasses().Lister().Get(name)
}

func (v *volumeSnapshotClassGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := v.informers.Snapshot().V1().VolumeSnapshotClasses().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, snapshotClass := range all {
		snapshotClass = snapshotClass.DeepCopy()
		count := v.countVolumeSnapshots(snapshotClass.Name)
		if snapshotClass.Annotations == nil {
			snapshotClass.Annotations = make(map[string]string)
		}
		snapshotClass.Annotations["kubesphere.io/snapshot-count"] = strconv.Itoa(count)
		result = append(result, snapshotClass)
	}

	return v1alpha3.DefaultList(result, query, v.compare, v.filter), nil
}

func (v *volumeSnapshotClassGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshotClass, ok := left.(*v1.VolumeSnapshotClass)
	if !ok {
		return false
	}
	rightSnapshotClass, ok := right.(*v1.VolumeSnapshotClass)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshotClass.ObjectMeta, rightSnapshotClass.ObjectMeta, field)
}

func (v *volumeSnapshotClassGetter) filter(object runtime.Object, filter query.Filter) bool {
	snapshotClass, ok := object.(*v1.VolumeSnapshotClass)
	if !ok {
		return false
	}

	switch filter.Field {
	case deletionPolicy:
		return strings.EqualFold(string(snapshotClass.DeletionPolicy), string(filter.Value))
	case driver:
		return strings.EqualFold(snapshotClass.Driver, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(snapshotClass.ObjectMeta, filter)
	}
}

func (v *volumeSnapshotClassGetter) countVolumeSnapshots(name string) int {
	snapshots, err := v.informers.Snapshot().V1().VolumeSnapshots().Lister().List(labels.Everything())
	if err != nil {
		return 0
	}
	var count int

	for _, snapshot := range snapshots {
		if snapshot.Spec.VolumeSnapshotClassName != nil && *snapshot.Spec.VolumeSnapshotClassName == name {
			count++
		}
	}
	return count
}
