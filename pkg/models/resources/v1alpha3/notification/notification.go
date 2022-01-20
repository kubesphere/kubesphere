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

package notification

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	notificationv2beta1 "kubesphere.io/api/notification/v2beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

func init() {
	crds.Filters[notificationv2beta1.SchemeGroupVersion.WithKind(notificationv2beta1.ResourceKindConfig)] = filter
	crds.Filters[notificationv2beta1.SchemeGroupVersion.WithKind(notificationv2beta1.ResourceKindReceiver)] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {

	accessor, err := meta.Accessor(object)
	if err != nil {
		return false
	}

	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if accessor.GetName() == name {
				return true
			}
		}
		return false
	case query.FieldName:
		return strings.Contains(accessor.GetName(), string(filter.Value))
	default:
		return true
	}
}
