package networkpolicy

import (
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type networkpolicyGetter struct {
	informers informers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory) v1alpha3.Interface {
	return &networkpolicyGetter{informers: informers}
}

func (n networkpolicyGetter) Get(namespace, name string) (runtime.Object, error) {
	return n.informers.Networking().V1().NetworkPolicies().Lister().NetworkPolicies(namespace).Get(name)
}

func (n networkpolicyGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	nps, err := n.informers.Networking().V1().NetworkPolicies().Lister().NetworkPolicies(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range nps {
		result = append(result, item)
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n networkpolicyGetter) filter(item runtime.Object, filter query.Filter) bool {
	np, ok := item.(*v1.NetworkPolicy)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(np.ObjectMeta, filter)
}

func (n networkpolicyGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNP, ok := left.(*v1.NetworkPolicy)
	if !ok {
		return false
	}

	rightNP, ok := right.(*v1.NetworkPolicy)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftNP.ObjectMeta, rightNP.ObjectMeta, field)
}
