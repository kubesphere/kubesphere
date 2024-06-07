/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package role

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type rolesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &rolesGetter{cache: cache}
}

func (d *rolesGetter) Get(namespace, name string) (runtime.Object, error) {
	role := &rbacv1.Role{}
	return role, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, role)
}

func (d *rolesGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	var roles []*rbacv1.Role
	var err error

	if aggregateTo := query.Filters[iamv1beta1.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(namespace, string(aggregateTo))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.AggregateTo)
	} else {
		roleList := &rbacv1.RoleList{}
		if err := d.cache.List(context.Background(), roleList, client.InNamespace(namespace),
			client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
			return nil, err
		}
		roles = make([]*rbacv1.Role, 0)
		for _, item := range roleList.Items {
			roles = append(roles, item.DeepCopy())
		}
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

	if annotation := obj.(*rbacv1.Role).Annotations[iamv1beta1.AggregationRolesAnnotation]; annotation != "" {
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
