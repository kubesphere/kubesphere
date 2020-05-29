/*
Copyright 2019 The KubeSphere Authors.

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

package clusterrole

import (
	"encoding/json"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type clusterrolesGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &clusterrolesGetter{sharedInformers: sharedInformers}
}

func (d *clusterrolesGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Rbac().V1().ClusterRoles().Lister().Get(name)
}

func (d *clusterrolesGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {

	var roles []*rbacv1.ClusterRole
	var err error

	if aggregateTo := query.Filters[iamv1alpha2.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(string(aggregateTo))
		delete(query.Filters, iamv1alpha2.AggregateTo)
	} else {
		roles, err = d.sharedInformers.Rbac().V1().ClusterRoles().Lister().List(query.Selector())
	}

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, clusterrole := range roles {
		result = append(result, clusterrole)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *clusterrolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftClusterRole, ok := left.(*rbacv1.ClusterRole)
	if !ok {
		return false
	}

	rightClusterRole, ok := right.(*rbacv1.ClusterRole)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftClusterRole.ObjectMeta, rightClusterRole.ObjectMeta, field)
}

func (d *clusterrolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.ClusterRole)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}

func (d *clusterrolesGetter) fetchAggregationRoles(name string) ([]*rbacv1.ClusterRole, error) {
	roles := make([]*rbacv1.ClusterRole, 0)

	obj, err := d.Get("", name)

	if err != nil {
		if errors.IsNotFound(err) {
			return roles, nil
		}
		return nil, err
	}

	if annotation := obj.(*rbacv1.ClusterRole).Annotations[iamv1alpha2.AggregationRolesAnnotation]; annotation != "" {
		var roleNames []string
		if err = json.Unmarshal([]byte(annotation), &roleNames); err == nil {

			for _, roleName := range roleNames {
				role, err := d.Get("", roleName)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("invalid aggregation role found: %s, %s", name, roleName)
						continue
					}
					return nil, err
				}

				roles = append(roles, role.(*rbacv1.ClusterRole))
			}
		}
	}

	return roles, nil
}
