/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package tenant

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

func (t *tenantOperator) CreateWorkspaceResourceQuota(workspace string, quota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error) {
	if quota.Labels == nil {
		quota.Labels = make(map[string]string)
	}
	quota.Labels[tenantv1beta1.WorkspaceLabel] = workspace
	quota.Spec.LabelSelector = labels.Set{tenantv1beta1.WorkspaceLabel: workspace}
	return quota, t.client.Create(context.Background(), quota)
}

func (t *tenantOperator) UpdateWorkspaceResourceQuota(workspace string, quota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error) {
	resourceQuota := &quotav1alpha2.ResourceQuota{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: quota.Name}, resourceQuota); err != nil {
		return nil, err
	}
	if resourceQuota.Labels[tenantv1beta1.WorkspaceLabel] != workspace {
		return nil, errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuota.Name)
	}
	quota = quota.DeepCopy()
	if quota.Labels == nil {
		quota.Labels = make(map[string]string)
	}
	quota.Labels[tenantv1beta1.WorkspaceLabel] = workspace
	quota.Spec.LabelSelector = labels.Set{tenantv1beta1.WorkspaceLabel: workspace}
	return quota, t.client.Update(context.Background(), quota)
}

func (t *tenantOperator) DeleteWorkspaceResourceQuota(workspace string, resourceQuotaName string) error {
	resourceQuota := &quotav1alpha2.ResourceQuota{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: resourceQuotaName}, resourceQuota); err != nil {
		return err
	}
	if resourceQuota.Labels[tenantv1beta1.WorkspaceLabel] != workspace {
		return errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuotaName)
	}
	return t.client.Delete(context.Background(), resourceQuota)
}

func (t *tenantOperator) DescribeWorkspaceResourceQuota(workspace string, resourceQuotaName string) (*quotav1alpha2.ResourceQuota, error) {
	resourceQuota := &quotav1alpha2.ResourceQuota{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: resourceQuotaName}, resourceQuota); err != nil {
		return nil, err
	}
	if resourceQuota.Labels[tenantv1beta1.WorkspaceLabel] != workspace {
		return nil, errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuotaName)
	}
	return resourceQuota, nil
}
