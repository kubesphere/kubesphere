/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package globalrole

import (
	"context"
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type globalRolesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &globalRolesGetter{cache: cache}
}

func (d *globalRolesGetter) Get(_, name string) (runtime.Object, error) {
	globalRole := &iamv1beta1.GlobalRole{}
	return globalRole, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, globalRole)
}

func (d *globalRolesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var roles []*iamv1beta1.GlobalRole
	var err error

	if aggregateTo := query.Filters[iamv1beta1.AggregateTo]; aggregateTo != "" {
		roles, err = d.fetchAggregationRoles(string(aggregateTo))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.AggregateTo)
	} else {
		globalRoleList := &iamv1beta1.GlobalRoleList{}
		if err := d.cache.List(context.Background(), globalRoleList,
			client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
			return nil, err
		}
		roles = make([]*iamv1beta1.GlobalRole, 0)
		for _, item := range globalRoleList.Items {
			roles = append(roles, item.DeepCopy())
		}
	}

	var result []runtime.Object
	for _, role := range roles {
		result = append(result, role)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *globalRolesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftRole, ok := left.(*iamv1beta1.GlobalRole)
	if !ok {
		return false
	}

	rightRole, ok := right.(*iamv1beta1.GlobalRole)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRole.ObjectMeta, rightRole.ObjectMeta, field)
}

func (d *globalRolesGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*iamv1beta1.GlobalRole)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}

func (d *globalRolesGetter) fetchAggregationRoles(name string) ([]*iamv1beta1.GlobalRole, error) {
	roles := make([]*iamv1beta1.GlobalRole, 0)

	obj, err := d.Get("", name)

	if err != nil {
		if errors.IsNotFound(err) {
			return roles, nil
		}
		return nil, err
	}

	if annotation := obj.(*iamv1beta1.GlobalRole).Annotations[iamv1beta1.AggregationRolesAnnotation]; annotation != "" {
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

				roles = append(roles, role.(*iamv1beta1.GlobalRole))
			}
		}
	}

	return roles, nil
}
