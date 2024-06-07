/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package statefulset

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type statefulSetGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &statefulSetGetter{cache: cache}
}

func (d *statefulSetGetter) Get(namespace, name string) (runtime.Object, error) {
	statefulSet := &appsv1.StatefulSet{}
	return statefulSet, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, statefulSet)
}

func (d *statefulSetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	statefulSets := &appsv1.StatefulSetList{}
	if err := d.cache.List(context.Background(), statefulSets, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range statefulSets.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *statefulSetGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftStatefulSet, ok := left.(*appsv1.StatefulSet)
	if !ok {
		return false
	}

	rightStatefulSet, ok := right.(*appsv1.StatefulSet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftStatefulSet.ObjectMeta, rightStatefulSet.ObjectMeta, field)
}

func (d *statefulSetGetter) filter(object runtime.Object, filter query.Filter) bool {
	statefulSet, ok := object.(*appsv1.StatefulSet)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return statefulSetStatus(statefulSet) == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(statefulSet.ObjectMeta, filter)
	}

}

func statefulSetStatus(item *appsv1.StatefulSet) string {
	if item.Spec.Replicas != nil {
		if item.Status.ReadyReplicas == 0 && *item.Spec.Replicas == 0 {
			return statusStopped
		} else if item.Status.ReadyReplicas == *item.Spec.Replicas {
			return statusRunning
		} else {
			return statusUpdating
		}
	}
	return statusStopped
}
