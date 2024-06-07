/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package customresourcedefinition

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type crdGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &crdGetter{cache: cache}
}

func (c crdGetter) Get(_, name string) (runtime.Object, error) {
	crd := &v1.CustomResourceDefinition{}
	return crd, c.cache.Get(context.Background(), types.NamespacedName{Name: name}, crd)
}

func (c crdGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	crds := &v1.CustomResourceDefinitionList{}
	if err := c.cache.List(context.Background(), crds,
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range crds.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
}

func (c crdGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftCRD, ok := left.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	rightCRD, ok := right.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCRD.ObjectMeta, rightCRD.ObjectMeta, field)
}

func (c crdGetter) filter(object runtime.Object, filter query.Filter) bool {
	crd, ok := object.(*v1.CustomResourceDefinition)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(crd.Name, string(filter.Value)) || strings.Contains(crd.Spec.Names.Kind, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(crd.ObjectMeta, filter)
	}
}
