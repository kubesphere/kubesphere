/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package predicate

import (
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"kubesphere.io/kubesphere/pkg/controller/cluster/utils"
)

type ClusterStatusChangedPredicate struct {
	predicate.Funcs
}

func (ClusterStatusChangedPredicate) Update(e event.UpdateEvent) bool {
	oldCluster, ok := e.ObjectOld.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}
	newCluster := e.ObjectNew.(*clusterv1alpha1.Cluster)
	// cluster is ready
	if !utils.IsClusterReady(oldCluster) && utils.IsClusterReady(newCluster) {
		return true
	}
	if !utils.IsClusterSchedulable(oldCluster) && utils.IsClusterSchedulable(newCluster) {
		return true
	}
	return false
}

func (ClusterStatusChangedPredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (ClusterStatusChangedPredicate) Delete(_ event.DeleteEvent) bool {
	return false
}

func (ClusterStatusChangedPredicate) Generic(_ event.GenericEvent) bool {
	return false
}
