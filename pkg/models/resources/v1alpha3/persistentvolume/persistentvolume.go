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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	storageClassName = "storageClassName"
)

func init() {
	crds.Filters[schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumes"}] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
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
		return crds.DefaultObjectMetaFilter(pv, filter)
	}
}
