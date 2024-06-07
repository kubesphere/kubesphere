/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package group

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

type groupGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &groupGetter{cache: cache}
}

func (d *groupGetter) Get(_, name string) (runtime.Object, error) {
	group := &iamv1beta1.Group{}
	return group, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, group)
}

func (d *groupGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	groups := &iamv1beta1.GroupList{}
	if err := d.cache.List(context.Background(), groups,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range groups.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *groupGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftGroup, ok := left.(*iamv1beta1.Group)
	if !ok {
		return false
	}

	rightGroup, ok := right.(*iamv1beta1.Group)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftGroup.ObjectMeta, rightGroup.ObjectMeta, field)
}

func (d *groupGetter) filter(object runtime.Object, filter query.Filter) bool {
	group, ok := object.(*iamv1beta1.Group)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(group.ObjectMeta, filter)
}
