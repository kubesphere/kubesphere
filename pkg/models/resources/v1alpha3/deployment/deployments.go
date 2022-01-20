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

package deployment

import (
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"

	v1 "k8s.io/api/apps/v1"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

func init() {
	crds.Filters[schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}] = filter
	crds.Comparers[schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}] = compare
}

func compare(left, right metav1.Object, field query.Field) bool {

	leftDeployment, ok := left.(*v1.Deployment)
	if !ok {
		return false
	}

	rightDeployment, ok := right.(*v1.Deployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeployment).After(lastUpdateTime(rightDeployment))
	default:
		return crds.DefaultObjectMetaCompare(leftDeployment, rightDeployment, field)
	}
}

func filter(object metav1.Object, filter query.Filter) bool {
	deployment, ok := object.(*v1.Deployment)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(deploymentStatus(deployment.Status), string(filter.Value)) == 0
	default:
		return crds.DefaultObjectMetaFilter(deployment, filter)
	}
}

func deploymentStatus(status v1.DeploymentStatus) string {
	if status.ReadyReplicas == 0 && status.Replicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}

func lastUpdateTime(deployment *v1.Deployment) time.Time {
	lut := deployment.CreationTimestamp.Time
	for _, condition := range deployment.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
