/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"kubesphere.io/api/cluster/v1alpha1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

func WorkspaceTemplateMatchTargetCluster(workspaceTemplate *tenantv1beta1.WorkspaceTemplate, cluster *v1alpha1.Cluster) bool {
	match := false
	if len(workspaceTemplate.Spec.Placement.Clusters) > 0 {
		for _, clusterRef := range workspaceTemplate.Spec.Placement.Clusters {
			if clusterRef.Name == cluster.Name {
				match = true
				break
			}
		}
	} else if workspaceTemplate.Spec.Placement.ClusterSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(workspaceTemplate.Spec.Placement.ClusterSelector)
		if err != nil {
			klog.Errorf("failed to parse cluster selector: %s", err)
			return false
		}
		match = selector.Matches(labels.Set(cluster.Labels))
	}
	return match
}
