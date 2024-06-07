/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacerolebinding

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const RoleName = "rolename"

type workspaceRoleBindingsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &workspaceRoleBindingsGetter{cache: cache}
}

func (d *workspaceRoleBindingsGetter) Get(_, name string) (runtime.Object, error) {
	workspaceRoleBinding := &iamv1beta1.WorkspaceRoleBinding{}
	return workspaceRoleBinding, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, workspaceRoleBinding)
}

func (d *workspaceRoleBindingsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	workspaceRoleBindings := &iamv1beta1.WorkspaceRoleBindingList{}
	if err := d.cache.List(context.Background(), workspaceRoleBindings,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range workspaceRoleBindings.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *workspaceRoleBindingsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftRoleBinding, ok := left.(*iamv1beta1.WorkspaceRoleBinding)
	if !ok {
		return false
	}

	rightRoleBinding, ok := right.(*iamv1beta1.WorkspaceRoleBinding)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRoleBinding.ObjectMeta, rightRoleBinding.ObjectMeta, field)
}

func (d *workspaceRoleBindingsGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*iamv1beta1.WorkspaceRoleBinding)

	if !ok {
		return false
	}
	switch filter.Field {
	case RoleName:
		return role.RoleRef.Name == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
	}
}
