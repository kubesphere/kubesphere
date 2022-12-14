package v1beta1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceCache struct {
	cache cache.Cache
}

func NewResourceCache(cache cache.Cache) Interface {
	return &resourceCache{cache: cache}
}

func (u *resourceCache) Get(name, namespace string, object client.Object) error {
	return u.cache.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, object)
}

func (u *resourceCache) List(namespace string, query *query.Query, list client.ObjectList) error {
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

func (u *resourceCache) compare(left, right metav1.Object, field query.Field) bool {
	return DefaultObjectMetaCompare(left, right, field)
}

func (u *resourceCache) filter(object metav1.Object, filter query.Filter) bool {
	return DefaultObjectMetaFilter(object, filter)
}
