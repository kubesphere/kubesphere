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

package clusterdashboard

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	monitoringdashboardv1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type dashboardGetter struct {
	c client.Reader
}

func New(c client.Reader) v1alpha3.Interface {
	return &dashboardGetter{c}
}

func (d *dashboardGetter) Get(_, name string) (runtime.Object, error) {
	dashboard := monitoringdashboardv1alpha2.ClusterDashboard{}
	err := d.c.Get(context.Background(), types.NamespacedName{Name: name}, &dashboard)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &dashboard, nil
}

func (d *dashboardGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	dashboards := monitoringdashboardv1alpha2.ClusterDashboardList{}
	err := d.c.List(context.Background(), &dashboards, &client.ListOptions{LabelSelector: query.Selector()})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var result []runtime.Object
	for i := range dashboards.Items {
		result = append(result, &dashboards.Items[i])
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *dashboardGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftClusterDashboard, ok := left.(*monitoringdashboardv1alpha2.ClusterDashboard)
	if !ok {
		return false
	}

	rightClusterDashboard, ok := right.(*monitoringdashboardv1alpha2.ClusterDashboard)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftClusterDashboard.ObjectMeta, rightClusterDashboard.ObjectMeta, field)
}

func (d *dashboardGetter) filter(object runtime.Object, filter query.Filter) bool {
	dashboard, ok := object.(*monitoringdashboardv1alpha2.ClusterDashboard)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(dashboard.ObjectMeta, filter)
}
