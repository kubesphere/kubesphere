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
	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	statusCreating = "creating"
	statusReady    = "ready"
	statusDeleting = "deleting"

	volumeSnapshotClassName   = "volumeSnapshotClassName"
	persistentVolumeClaimName = "persistentVolumeClaimName"
)

func init() {
	crds.Filters[v1.SchemeGroupVersion.WithKind("VolumeSnapshot")] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
	snapshot, ok := object.(*v1.VolumeSnapshot)
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
		return crds.DefaultObjectMetaFilter(snapshot, filter)
	}
}

func snapshotStatus(item *v1.VolumeSnapshot) string {
	status := statusCreating
	if item != nil && item.Status != nil && item.Status.ReadyToUse != nil && *item.Status.ReadyToUse {
		status = statusReady
	}
	if item != nil && item.DeletionTimestamp != nil {
		status = statusDeleting
	}
	return status
}
