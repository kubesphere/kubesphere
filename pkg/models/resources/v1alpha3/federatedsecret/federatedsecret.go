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

package federatedsecret

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

func init() {
	crds.Filters[v1beta1.SchemeGroupVersion.WithKind(v1beta1.FederatedSecretKind)] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
	fedSecret, ok := object.(*v1beta1.FederatedSecret)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldType:
		return strings.Compare(string(fedSecret.Spec.Template.Type), string(filter.Value)) == 0
	default:
		return crds.DefaultObjectMetaFilter(fedSecret, filter)
	}
}
