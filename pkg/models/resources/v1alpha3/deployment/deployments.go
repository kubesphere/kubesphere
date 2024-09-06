/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package deployment

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"

	appsv1 "k8s.io/api/apps/v1"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type deploymentsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &deploymentsGetter{cache: cache}
}

func (d *deploymentsGetter) Get(namespace, name string) (runtime.Object, error) {
	deployment := &appsv1.Deployment{}
	return deployment, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, deployment)
}

func (d *deploymentsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	deployments := &appsv1.DeploymentList{}
	if err := d.cache.List(context.Background(), deployments, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range deployments.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *deploymentsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftDeployment, ok := left.(*appsv1.Deployment)
	if !ok {
		return false
	}

	rightDeployment, ok := right.(*appsv1.Deployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeployment).After(lastUpdateTime(rightDeployment))
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftDeployment.ObjectMeta, rightDeployment.ObjectMeta, field)
	}
}

func (d *deploymentsGetter) filter(object runtime.Object, filter query.Filter) bool {
	deployment, ok := object.(*appsv1.Deployment)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(deploymentStatus(deployment.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(deployment.ObjectMeta, filter)
	}
}

func deploymentStatus(status appsv1.DeploymentStatus) string {
	if status.ReadyReplicas == 0 && status.Replicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}

func lastUpdateTime(deployment *appsv1.Deployment) time.Time {
	lut := deployment.CreationTimestamp.Time
	for _, condition := range deployment.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
