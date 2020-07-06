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

package volumesnapshot

import (
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	statusCreating = "creating"
	statusReady    = "ready"

	volumeSnapshotClassName   = "volumeSnapshotClassName"
	persistentVolumeClaimName = "persistentVolumeClaimName"
)

type volumeSnapshotGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informer externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &volumeSnapshotGetter{informers: informer}
}

func (v *volumeSnapshotGetter) Get(namespace, name string) (runtime.Object, error) {
	return v.informers.Snapshot().V1beta1().VolumeSnapshots().Lister().VolumeSnapshots(namespace).Get(name)
}

func (v *volumeSnapshotGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := v.informers.Snapshot().V1beta1().VolumeSnapshots().Lister().VolumeSnapshots(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, snapshot := range all {
		result = append(result, snapshot)
	}

	return v1alpha3.DefaultList(result, query, v.compare, v.filter), nil
}

func (v *volumeSnapshotGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshot, ok := left.(*v1beta1.VolumeSnapshot)
	if !ok {
		return false
	}
	rightSnapshot, ok := right.(*v1beta1.VolumeSnapshot)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshot.ObjectMeta, rightSnapshot.ObjectMeta, field)
}

func (v *volumeSnapshotGetter) filter(object runtime.Object, filter query.Filter) bool {
	snapshot, ok := object.(*v1beta1.VolumeSnapshot)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return snapshotStatus(snapshot) == string(filter.Value)
	case volumeSnapshotClassName:
		name := snapshot.Spec.VolumeSnapshotClassName
		return name != nil && *name == string(filter.Value)
	case persistentVolumeClaimName:
		name := snapshot.Spec.Source.PersistentVolumeClaimName
		return name != nil && *name == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(snapshot.ObjectMeta, filter)
	}
}

func snapshotStatus(item *v1beta1.VolumeSnapshot) string {
	status := statusCreating
	if item != nil && item.Status != nil && item.Status.ReadyToUse != nil && *item.Status.ReadyToUse {
		status = statusReady
	}
	return status
}
