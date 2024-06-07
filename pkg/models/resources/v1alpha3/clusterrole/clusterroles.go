/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package clusterrole

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

type clusterRolesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &clusterRolesGetter{cache: cache}
}

func (d *clusterRolesGetter) Get(_, name string) (runtime.Object, error) {
	clusterRole := &rbacv1.ClusterRole{}
	return clusterRole, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, clusterRole)
}

func (d *clusterRolesGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	var roles []*rbacv1.ClusterRole
	var err error
	if aggregateTo := query.Filters[iamv1beta1.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(string(aggregateTo))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.AggregateTo)
	} else {
		clusterRoleList := &rbacv1.ClusterRoleList{}
		if err := d.cache.List(context.Background(), clusterRoleList,
			client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
			return nil, err
		}
		roles = make([]*rbacv1.ClusterRole, 0)
		for _, item := range clusterRoleList.Items {
			roles = append(roles, item.DeepCopy())
		}
	}

	var result []runtime.Object
	for _, clusterRole := range roles {
		result = append(result, clusterRole)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *clusterRolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
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

func (d *clusterRolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.ClusterRole)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}

func (d *clusterRolesGetter) fetchAggregationRoles(name string) ([]*rbacv1.ClusterRole, error) {
	roles := make([]*rbacv1.ClusterRole, 0)

	obj, err := d.Get("", name)

	if err != nil {
		if errors.IsNotFound(err) {
			return roles, nil
		}
		return nil, err
	}

	if annotation := obj.(*rbacv1.ClusterRole).Annotations[iamv1beta1.AggregationRolesAnnotation]; annotation != "" {
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
