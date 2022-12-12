package reader

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reader struct {
	cache cache.Cache
}

func NewReader(cache cache.Cache) Reader {
	return &reader{cache: cache}
}

func (u *reader) Get(namespace, name string, object client.Object) error {
	return u.cache.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, object)
}

func (u *reader) List(namespace string, query *query.Query, list client.ObjectList) error {
	listOpt := &client.ListOptions{
		LabelSelector: query.Selector(),
		Namespace:     namespace,
	}
	err := u.cache.List(context.Background(), list, listOpt)
	if err != nil {
		return err
	}

	DefaultList(list, query, u.compare, u.filter)
	return nil
}

func (u *reader) compare(left, right metav1.Object, field query.Field) bool {
	return DefaultObjectMetaCompare(left, right, field)
}

func (u *reader) filter(object metav1.Object, filter query.Filter) bool {
	return DefaultObjectMetaFilter(object, filter)
}
