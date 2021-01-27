/*

 Copyright 2021 The KubeSphere Authors.

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

package tenant

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	quotav1alpha2 "kubesphere.io/kubesphere/pkg/apis/quota/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
)

func (t *tenantOperator) CreateWorkspaceResourceQuota(workspace string, quota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error) {
	if quota.Labels == nil {
		quota.Labels = make(map[string]string)
	}
	quota.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	quota.Spec.LabelSelector = labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}
	return t.ksclient.QuotaV1alpha2().ResourceQuotas().Create(context.TODO(), quota, metav1.CreateOptions{})
}

func (t *tenantOperator) UpdateWorkspaceResourceQuota(workspace string, quota *quotav1alpha2.ResourceQuota) (*quotav1alpha2.ResourceQuota, error) {
	resourceQuota, err := t.ksclient.QuotaV1alpha2().ResourceQuotas().Get(context.TODO(), quota.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if resourceQuota.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		return nil, errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuota.Name)
	}
	if quota.Labels == nil {
		quota.Labels = make(map[string]string)
	}
	quota.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	quota.Spec.LabelSelector = labels.Set{tenantv1alpha1.WorkspaceLabel: workspace}
	return t.ksclient.QuotaV1alpha2().ResourceQuotas().Update(context.TODO(), quota, metav1.UpdateOptions{})
}

func (t *tenantOperator) DeleteWorkspaceResourceQuota(workspace string, resourceQuotaName string) error {
	resourceQuota, err := t.ksclient.QuotaV1alpha2().ResourceQuotas().Get(context.TODO(), resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if resourceQuota.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		return errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuotaName)
	}
	return t.ksclient.QuotaV1alpha2().ResourceQuotas().Delete(context.TODO(), resourceQuotaName, metav1.DeleteOptions{})
}

func (t *tenantOperator) DescribeWorkspaceResourceQuota(workspace string, resourceQuotaName string) (*quotav1alpha2.ResourceQuota, error) {
	resourceQuota, err := t.ksclient.QuotaV1alpha2().ResourceQuotas().Get(context.TODO(), resourceQuotaName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if resourceQuota.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		return nil, errors.NewNotFound(quotav1alpha2.Resource(quotav1alpha2.ResourcesSingularCluster), resourceQuotaName)
	}
	return resourceQuota, nil
}
