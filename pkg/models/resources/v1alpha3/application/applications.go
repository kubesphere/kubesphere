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

package application

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	appv1beta1 "sigs.k8s.io/application/api/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

func init() {
	crds.Comparers[schema.GroupVersionKind{Group: "app.k8s.io", Version: "v1beta1", Kind: "Application"}] = compare
}

func compare(left, right metav1.Object, field query.Field) bool {

	leftApplication, ok := left.(*appv1beta1.Application)
	if !ok {
		return false
	}

	rightApplication, ok := right.(*appv1beta1.Application)
	if !ok {
		return false
	}
	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftApplication).After(lastUpdateTime(rightApplication))
	default:
		return crds.DefaultObjectMetaCompare(leftApplication, rightApplication, field)
	}
}

func lastUpdateTime(application *appv1beta1.Application) time.Time {
	lut := application.CreationTimestamp.Time
	for _, condition := range application.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
