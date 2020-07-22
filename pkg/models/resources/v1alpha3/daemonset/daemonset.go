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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type daemonSetGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &daemonSetGetter{sharedInformers: sharedInformers}
}

func (d *daemonSetGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)
}

func (d *daemonSetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	// first retrieves all daemonSets within given namespace
	daemonSets, err := d.sharedInformers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, daemonSet := range daemonSets {
		result = append(result, daemonSet)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *daemonSetGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftDaemonSet, ok := left.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	rightDaemonSet, ok := right.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftDaemonSet.ObjectMeta, rightDaemonSet.ObjectMeta, field)
}

func (d *daemonSetGetter) filter(object runtime.Object, filter query.Filter) bool {
	daemonSet, ok := object.(*appsv1.DaemonSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(daemonsetStatus(&daemonSet.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(daemonSet.ObjectMeta, filter)
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
