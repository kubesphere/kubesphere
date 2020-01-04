package namespace

import (
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/informers"
    "kubesphere.io/kubesphere/pkg/api"
    "kubesphere.io/kubesphere/pkg/apiserver/query"
    "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
    "strings"
)

type namespaceGetter struct {
    informers informers.SharedInformerFactory
}

func NewNamespaceGetter(informers informers.SharedInformerFactory) v1alpha3.Interface {
    return &namespaceGetter{informers:informers}
}

func (n namespaceGetter) Get(_, name string) (runtime.Object, error) {
    return n.informers.Core().V1().Namespaces().Lister().Get(name)
}

func (n namespaceGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
    ns, err := n.informers.Core().V1().Namespaces().Lister().List(labels.Everything())
    if err != nil {
        return nil, err
    }

    var result []runtime.Object
    for _, item := range ns {
        result = append(result, item)
    }

    return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n namespaceGetter) filter(item runtime.Object, filter query.Filter) bool {
    namespace, ok := item.(*v1.Namespace)
    if !ok {
        return false
    }

    switch filter.Field {
    case query.FieldName:
        return query.ComparableString(namespace.Name).Contains(filter.Value)
    case query.FieldStatus:
        return query.ComparableString(namespace.Status.Phase).Compare(filter.Value) == 0
    default:
        return false
    }
}

func (n namespaceGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
    leftNs, ok := left.(*v1.Namespace)
    if !ok {
        return false
    }

    rightNs, ok := right.(*v1.Namespace)
    if !ok {
        return true
    }

    switch field {
    case query.FieldName:
        return strings.Compare(leftNs.Name, rightNs.Name) > 0
    case query.FieldCreationTimeStamp:
        return leftNs.CreationTimestamp.After(rightNs.CreationTimestamp.Time)
    default:
        return false
    }
}