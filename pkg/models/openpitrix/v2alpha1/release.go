/*
Copyright 2020 The KubeSphere Authors.

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

package v2alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/openpitrix/helmrelease"
)

type HelmReleaseInterface interface {
	DescribeApplication(workspace, clusterName, namespace, applicationId string) (*v1alpha1.HelmRelease, error)
	ListApplications(workspace, cluster, namespace string, q *query.Query) (*api.ListResult, error)
}
type releaseOperator struct {
	rlsGetter resources.Interface
}

func newReleaseOperator(ksFactory externalversions.SharedInformerFactory) HelmReleaseInterface {
	c := &releaseOperator{
		rlsGetter: helmrelease.New(ksFactory),
	}

	return c
}
func (c *releaseOperator) DescribeApplication(workspace, clusterName, namespace, applicationId string) (*v1alpha1.HelmRelease, error) {
	ret, err := c.rlsGetter.Get("", applicationId)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	rls := ret.(*v1alpha1.HelmRelease)
	return rls, nil
}

func (c *releaseOperator) ListApplications(workspace, cluster, namespace string, q *query.Query) (*api.ListResult, error) {

	labelSelector, err := labels.ConvertSelectorToLabelsMap(q.LabelSelector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	extra := labels.Set{}
	if workspace != "" {
		extra[constants.WorkspaceLabelKey] = workspace
	}

	// cluster must used with namespace
	if cluster != "" {
		extra[constants.ClusterNameLabelKey] = cluster
	}
	if namespace != "" {
		extra[constants.NamespaceLabelKey] = namespace
	}
	if len(extra) > 0 {
		q.LabelSelector = labels.Merge(labelSelector, extra).String()
	}

	releases, err := c.rlsGetter.List("", q)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list app release failed, error: %v", err)
		return nil, err
	}

	return releases, nil
}
