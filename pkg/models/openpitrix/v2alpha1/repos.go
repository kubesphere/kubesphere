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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/openpitrix/repo"
)

type RepoInterface interface {
	ListRepos(workspace string, q *query.Query) (*api.ListResult, error)
	DescribeRepo(id string) (*v1alpha1.HelmRepo, error)
}

type repoOperator struct {
	reposGetter resources.Interface
}

func newRepoOperator(factory externalversions.SharedInformerFactory) RepoInterface {
	return &repoOperator{
		reposGetter: repo.New(factory),
	}
}

func (c *repoOperator) DescribeRepo(id string) (*v1alpha1.HelmRepo, error) {
	result, err := c.reposGetter.Get("", id)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	repo := result.(*v1alpha1.HelmRepo)
	repo.Status.Data = ""

	return repo, nil
}

func (c *repoOperator) ListRepos(workspace string, qry *query.Query) (result *api.ListResult, err error) {
	if workspace != "" {
		labelSelector, err := labels.ConvertSelectorToLabelsMap(qry.LabelSelector)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		qry.LabelSelector = labels.Merge(labelSelector, labels.Set{constants.WorkspaceLabelKey: workspace}).String()
	}
	result, err = c.reposGetter.List("", qry)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// remove status data and credential
	for i := range result.Items {
		d := result.Items[i].(*v1alpha1.HelmRepo)
		d.Status.Data = ""
		d.Spec.Credential = v1alpha1.HelmRepoCredential{}
	}

	return result, nil
}
