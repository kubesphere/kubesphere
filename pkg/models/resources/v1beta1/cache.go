package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

type resourceCache struct {
	cache cache.Cache
}

func NewResourceCache(cache cache.Cache) Interface {
	return &resourceCache{cache: cache}
}

func (u *resourceCache) Get(namespace, name string, object client.Object) error {
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

	extractList, err := meta.ExtractList(list)
	if err != nil {
		return err
	}

	filtered := DefaultList(extractList, query, compare, filter)
	if err := meta.SetList(list, filtered); err != nil {
		return err
	}
	return nil
}

func compare(left, right runtime.Object, field query.Field) bool {
	l, err := meta.Accessor(left)
	if err != nil {
		return false
	}
	r, err := meta.Accessor(right)
	if err != nil {
		return false
	}
	return DefaultObjectMetaCompare(l, r, field)
}

func filter(object runtime.Object, filter query.Filter) bool {
	o, err := meta.Accessor(object)
	if err != nil {
		return false
	}
	return DefaultObjectMetaFilter(o, filter)
}
