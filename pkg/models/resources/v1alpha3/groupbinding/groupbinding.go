/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package groupbinding

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
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const User = "user"

type groupBindingGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &groupBindingGetter{cache: cache}
}

func (d *groupBindingGetter) Get(_, name string) (runtime.Object, error) {
	groupBinding := &iamv1beta1.GroupBinding{}
	return groupBinding, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, groupBinding)
}

func (d *groupBindingGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	groupBindings := &iamv1beta1.GroupBindingList{}
	if err := d.cache.List(context.Background(), groupBindings,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range groupBindings.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *groupBindingGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftGroupBinding, ok := left.(*iamv1beta1.GroupBinding)
	if !ok {
		return false
	}

	rightGroupBinding, ok := right.(*iamv1beta1.GroupBinding)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftGroupBinding.ObjectMeta, rightGroupBinding.ObjectMeta, field)
}

func (d *groupBindingGetter) filter(object runtime.Object, filter query.Filter) bool {
	groupBinding, ok := object.(*iamv1beta1.GroupBinding)
	if !ok {
		return false
	}

	switch filter.Field {
	case User:
		return sliceutil.HasString(groupBinding.Users, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(groupBinding.ObjectMeta, filter)
	}
}
