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

package federatedapplication

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

func init() {
	crds.Comparers[v1beta1.SchemeGroupVersion.WithKind(v1beta1.FederatedApplicationKind)] = compare
}

func compare(left, right metav1.Object, field query.Field) bool {

	leftApplication, ok := left.(*v1beta1.FederatedApplication)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*v1beta1.FederatedApplication)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftApplication) > (lastUpdateTime(rightApplication))
	default:
		return crds.DefaultObjectMetaCompare(leftApplication, rightApplication, field)
	}
}

func lastUpdateTime(application *v1beta1.FederatedApplication) string {
	lut := application.CreationTimestamp.Time.String()
	for _, condition := range application.Status.Conditions {
		if condition.LastUpdateTime > lut {
			lut = condition.LastUpdateTime
		}
	}
	return lut
}
