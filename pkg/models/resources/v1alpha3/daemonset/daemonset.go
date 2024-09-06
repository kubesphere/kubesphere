/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package daemonset

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type daemonSetGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &daemonSetGetter{cache: cache}
}

func (d *daemonSetGetter) Get(namespace, name string) (runtime.Object, error) {
	daemonSet := &appsv1.DaemonSet{}
	return daemonSet, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, daemonSet)
}

func (d *daemonSetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	daemonSets := &appsv1.DaemonSetList{}
	if err := d.cache.List(context.Background(), daemonSets, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range daemonSets.Items {
		result = append(result, item.DeepCopy())
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
		return strings.Compare(daemonSetStatus(&daemonSet.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(daemonSet.ObjectMeta, filter)
	}
}

func daemonSetStatus(status *appsv1.DaemonSetStatus) string {
	if status.DesiredNumberScheduled == 0 && status.NumberReady == 0 {
		return statusStopped
	} else if status.DesiredNumberScheduled == status.NumberReady {
		return statusRunning
	} else {
		return statusUpdating
	}
}
