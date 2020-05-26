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

package workspacerole

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type workspacerolesGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &workspacerolesGetter{sharedInformers: sharedInformers}
}

func (d *workspacerolesGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.Iam().V1alpha2().WorkspaceRoles().Lister().Get(name)
}

func (d *workspacerolesGetter) List(_ string, queryParam *query.Query) (*api.ListResult, error) {

	var roles []*iamv1alpha2.WorkspaceRole
	var err error

	if aggregateTo := queryParam.Filters[iamv1alpha2.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(string(aggregateTo))
		delete(queryParam.Filters, iamv1alpha2.AggregateTo)
	} else {
		roles, err = d.sharedInformers.Iam().V1alpha2().WorkspaceRoles().Lister().List(queryParam.Selector())
	}

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, role := range roles {
		result = append(result, role)
	}

	return v1alpha3.DefaultList(result, queryParam, d.compare, d.filter), nil
}

func (d *workspacerolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftRole, ok := left.(*iamv1alpha2.WorkspaceRole)
	if !ok {
		return false
	}

	rightRole, ok := right.(*iamv1alpha2.WorkspaceRole)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRole.ObjectMeta, rightRole.ObjectMeta, field)
}

func (d *workspacerolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*iamv1alpha2.WorkspaceRole)

	if !ok {
		return false
	}

	switch filter.Field {
	case iamv1alpha2.ScopeWorkspace:
		return role.Labels[tenantv1alpha1.WorkspaceLabel] == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
	}

}

func (d *workspacerolesGetter) fetchAggregationRoles(name string) ([]*iamv1alpha2.WorkspaceRole, error) {
	roles := make([]*iamv1alpha2.WorkspaceRole, 0)

	obj, err := d.Get("", name)

	if err != nil {
		if errors.IsNotFound(err) {
			return roles, nil
		}
		return nil, err
	}

	if annotation := obj.(*iamv1alpha2.WorkspaceRole).Annotations[iamv1alpha2.AggregationRolesAnnotation]; annotation != "" {
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

				roles = append(roles, role.(*iamv1alpha2.WorkspaceRole))
			}
		}
	}

	return roles, nil
}
