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

package daemonset

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

func init() {
	crds.Filters[schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}] = filter
}

func filter(object metav1.Object, filter query.Filter) bool {
	daemonSet, ok := object.(*appsv1.DaemonSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(daemonsetStatus(&daemonSet.Status), string(filter.Value)) == 0
	default:
		return crds.DefaultObjectMetaFilter(daemonSet, filter)
	}
}

func daemonsetStatus(status *appsv1.DaemonSetStatus) string {
	if status.DesiredNumberScheduled == 0 && status.NumberReady == 0 {
		return statusStopped
	} else if status.DesiredNumberScheduled == status.NumberReady {
		return statusRunning
	} else {
		return statusUpdating
	}
}
