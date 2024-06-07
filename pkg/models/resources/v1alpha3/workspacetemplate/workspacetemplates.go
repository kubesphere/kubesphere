/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacetemplate

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type workspaceGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &workspaceGetter{cache: cache}
}

func (d *workspaceGetter) Get(_, name string) (runtime.Object, error) {
	workspaceTemplate := &tenantv1beta1.WorkspaceTemplate{}
	return workspaceTemplate, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, workspaceTemplate)
}

func (d *workspaceGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	workspaces := &tenantv1beta1.WorkspaceTemplateList{}
	if err := d.cache.List(context.Background(), workspaces,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range workspaces.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *workspaceGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftWorkspace, ok := left.(*tenantv1beta1.WorkspaceTemplate)
	if !ok {
		return false
	}

	rightWorkspace, ok := right.(*tenantv1beta1.WorkspaceTemplate)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftWorkspace.ObjectMeta, rightWorkspace.ObjectMeta, field)
}

func (d *workspaceGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*tenantv1beta1.WorkspaceTemplate)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
