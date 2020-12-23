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
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/openpitrix/application"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/openpitrix/applicationversion"
)

type ApplicationInterface interface {
	DescribeAppVersion(id string) (*v1alpha1.HelmApplicationVersion, error)
	DescribeApp(id string) (*v1alpha1.HelmApplication, error)

	ListApps(workspace string, q *query.Query) (*api.ListResult, error)
	ListAppVersions(workspace, appId string, q *query.Query) (*api.ListResult, error)
}

type applicationOperator struct {
	appsGetter       resources.Interface
	appVersionGetter resources.Interface
}

func newApplicationOperator(informers externalversions.SharedInformerFactory) ApplicationInterface {
	op := &applicationOperator{
		appsGetter:       application.New(informers),
		appVersionGetter: applicationversion.New(informers),
	}

	return op
}

func (c *applicationOperator) ListApps(workspace string, q *query.Query) (*api.ListResult, error) {

	labelSelector, err := labels.ConvertSelectorToLabelsMap(q.LabelSelector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	extra := labels.Set{}
	if workspace != "" {
		extra[constants.WorkspaceLabelKey] = workspace
	}

	if len(extra) > 0 {
		q.LabelSelector = labels.Merge(labelSelector, extra).String()
	}

	releases, err := c.appsGetter.List("", q)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list app failed, error: %v", err)
		return nil, err
	}

	return releases, nil
}

func (c *applicationOperator) DescribeApp(verId string) (*v1alpha1.HelmApplication, error) {
	ret, err := c.appsGetter.Get("", verId)
	if err != nil {
		klog.Errorf("get app failed, error: %v", err)
		return nil, err
	}

	return ret.(*v1alpha1.HelmApplication), nil
}
