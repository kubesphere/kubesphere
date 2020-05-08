package cluster

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type clustersGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informers externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &clustersGetter{
		informers: informers,
	}
}

func (c clustersGetter) Get(_, name string) (runtime.Object, error) {
	return c.informers.Cluster().V1alpha1().Clusters().Lister().Get(name)
}

func (c clustersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	clusters, err := c.informers.Cluster().V1alpha1().Clusters().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range clusters {
		result = append(result, deploy)
	}

	return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
}

func (c clustersGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
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

func (c clustersGetter) filter(object runtime.Object, filter query.Filter) bool {
	cluster, ok := object.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(cluster.ObjectMeta, filter)
}
