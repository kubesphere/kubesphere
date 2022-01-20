/*
Copyright 2020 KubeSphere Authors

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

package federateddeployment

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
	leftFedDeployment, ok := left.(*v1beta1.FederatedDeployment)
	if !ok {
		return false
	}

	rightFedDeployment, ok := right.(*v1beta1.FederatedDeployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftFedDeployment) > lastUpdateTime(rightFedDeployment)
	default:
		return crds.DefaultObjectMetaCompare(leftFedDeployment, rightFedDeployment, field)
	}
}

func lastUpdateTime(fedDeployment *v1beta1.FederatedDeployment) string {
	lut := fedDeployment.CreationTimestamp.Time.String()
	for _, condition := range fedDeployment.Status.Conditions {
		if condition.LastUpdateTime > lut {
			lut = condition.LastUpdateTime
		}
	}
	return lut
}
