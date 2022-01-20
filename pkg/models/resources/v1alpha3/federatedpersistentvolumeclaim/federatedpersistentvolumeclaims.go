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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	storageClassName = "storageClassName"
)

func init() {
	crds.Filters[v1beta1.SchemeGroupVersion.WithKind(v1beta1.FederatedPersistentVolumeClaimKind)] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
	pvc, ok := object.(*v1beta1.FederatedPersistentVolumeClaim)
	if !ok {
		return false
	}

	switch filter.Field {
	case storageClassName:
		return pvc.Spec.Template.Spec.StorageClassName != nil && *pvc.Spec.Template.Spec.StorageClassName == string(filter.Value)
	default:
		return crds.DefaultObjectMetaFilter(pvc, filter)
	}
}
