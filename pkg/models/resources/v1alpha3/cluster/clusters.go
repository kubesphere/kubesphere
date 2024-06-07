/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type clustersGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &clustersGetter{
		cache: cache,
	}
}

func (c *clustersGetter) Get(_, name string) (runtime.Object, error) {
	cluster := &clusterv1alpha1.Cluster{}
	if err := c.cache.Get(context.Background(), types.NamespacedName{Name: name}, cluster); err != nil {
		return nil, err
	}
	return c.transform(cluster), nil
}

func (c *clustersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	cluster := &clusterv1alpha1.ClusterList{}
	if err := c.cache.List(context.Background(), cluster, client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range cluster.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, c.compare, c.filter, c.transform), nil
}

func (c *clustersGetter) transform(obj runtime.Object) runtime.Object {
	in := obj.(*clusterv1alpha1.Cluster)
	out := in.DeepCopy()
	out.Spec.Connection.KubeConfig = nil
	return out
}

func (c *clustersGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftCluster, ok := left.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	rightCluster, ok := right.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCluster.ObjectMeta, rightCluster.ObjectMeta, field)
}

func (c *clustersGetter) filter(object runtime.Object, filter query.Filter) bool {
	cluster, ok := object.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(cluster.ObjectMeta, filter)
}
