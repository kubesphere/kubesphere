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

package role

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

type rolesGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &rolesGetter{sharedInformers: sharedInformers}
}

func (d *rolesGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Rbac().V1().Roles().Lister().Roles(namespace).Get(name)
}

func (d *rolesGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {

	var roles []*rbacv1.Role
	var err error

	if aggregateTo := query.Filters[iamv1alpha2.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(namespace, string(aggregateTo))
		delete(query.Filters, iamv1alpha2.AggregateTo)
	} else {
		roles, err = d.sharedInformers.Rbac().V1().Roles().Lister().Roles(namespace).List(query.Selector())
	}

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, role := range roles {
		result = append(result, role)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *rolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftRole, ok := left.(*rbacv1.Role)
	if !ok {
		return false
	}

	rightRole, ok := right.(*rbacv1.Role)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRole.ObjectMeta, rightRole.ObjectMeta, field)
}

func (d *rolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.Role)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}

func (d *rolesGetter) fetchAggregationRoles(namespace, name string) ([]*rbacv1.Role, error) {
	roles := make([]*rbacv1.Role, 0)

	obj, err := d.Get(namespace, name)

	if err != nil {
		if errors.IsNotFound(err) {
			return roles, nil
		}
		return nil, err
	}

	if annotation := obj.(*rbacv1.Role).Annotations[iamv1alpha2.AggregationRolesAnnotation]; annotation != "" {
		var roleNames []string
		if err = json.Unmarshal([]byte(annotation), &roleNames); err == nil {
			for _, roleName := range roleNames {
				role, err := d.Get(namespace, roleName)
				if err != nil {
					if errors.IsNotFound(err) {
						klog.V(6).Infof("invalid aggregation role found: %s, %s", name, roleName)
						continue
					}
					klog.Error(err)
					return nil, err
				}
				roles = append(roles, role.(*rbacv1.Role))
			}
		}
	}

	return roles, nil
}
