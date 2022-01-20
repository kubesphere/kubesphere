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
	"strings"

	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	deletionPolicy = "deletionPolicy"
	driver         = "driver"
)

func init() {
	crds.Filters[v1.SchemeGroupVersion.WithKind("VolumeSnapshotClass")] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
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
		return crds.DefaultObjectMetaFilter(snapshotClass, filter)
	}
}
