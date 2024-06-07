/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package clusterrolebinding

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type clusterRoleBindingsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &clusterRoleBindingsGetter{cache: cache}
}

func (d *clusterRoleBindingsGetter) Get(_, name string) (runtime.Object, error) {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	return clusterRoleBinding, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, clusterRoleBinding)
}

func (d *clusterRoleBindingsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	clusterRoleBindings := &rbacv1.ClusterRoleBindingList{}
	if err := d.cache.List(context.Background(), clusterRoleBindings,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range clusterRoleBindings.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *clusterRoleBindingsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftRoleBinding, ok := left.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return false
	}

	rightRoleBinding, ok := right.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRoleBinding.ObjectMeta, rightRoleBinding.ObjectMeta, field)
}

func (d *clusterRoleBindingsGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.ClusterRoleBinding)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
