/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package label

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type labelsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &labelsGetter{cache: cache}
}

func (n labelsGetter) Get(_, name string) (runtime.Object, error) {
	label := &clusterv1alpha1.Label{}
	return label, n.cache.Get(context.Background(), types.NamespacedName{Name: name}, label)
}

func (n labelsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	labels := &clusterv1alpha1.LabelList{}
	if err := n.cache.List(context.Background(), labels, client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range labels.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n labelsGetter) filter(item runtime.Object, filter query.Filter) bool {
	label, ok := item.(*clusterv1alpha1.Label)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(label.Spec.Key, string(filter.Value)) || strings.Contains(label.Spec.Value, string(filter.Value))
	default:
		return false
	}
}

func (n labelsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftLabel, ok := left.(*clusterv1alpha1.Label)
	if !ok {
		return false
	}

	rightLabel, ok := right.(*clusterv1alpha1.Label)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftLabel.ObjectMeta, rightLabel.ObjectMeta, field)
}
