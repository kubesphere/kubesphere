package customresourcedefinition

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type crdGetter struct {
	informers apiextensionsinformers.SharedInformerFactory
}

func New(informers apiextensionsinformers.SharedInformerFactory) v1alpha3.Interface {
	return &crdGetter{
		informers: informers,
	}
}

func (c crdGetter) Get(_, name string) (runtime.Object, error) {
	return c.informers.Apiextensions().V1().CustomResourceDefinitions().Lister().Get(name)
}

func (c crdGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	crds, err := c.informers.Apiextensions().V1().CustomResourceDefinitions().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, crd := range crds {
		result = append(result, crd)
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

	return v1alpha3.DefaultObjectMetaFilter(crd.ObjectMeta, filter)
}
