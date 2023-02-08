package v1beta1

import (
	"context"
	"errors"
	"strings"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	"kubesphere.io/kubesphere/pkg/apiserver/query"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrResourceNotSupported = errors.New("resource is not supported")
var ErrResourceNotServed = errors.New("resource is not served")

const labelResourceServed = "kubesphere.io/resource-served"

// TODO If delete the crd at the cluster when ks is running, the client.cache doesn`t return err but empty result
func New(client client.Client, cache cache.Cache) ResourceGetter {
	return &resourceGetter{
		client:   client,
		cache:    NewResourceCache(cache),
		serveCRD: make(map[string]bool, 0),
	}
}

type resourceGetter struct {
	client   client.Client
	cache    Interface
	serveCRD map[string]bool
	sync.RWMutex
}

func (h *resourceGetter) GetResource(gvr schema.GroupVersionResource, name, namespace string) (client.Object, error) {
	var obj client.Object
	gvk, err := h.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	if h.client.Scheme().Recognizes(gvk) {
		gvkObject, err := h.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	} else {
		serviced, err := h.isServed(gvr)
		if err != nil {
			return nil, err
		}
		if !serviced {
			return nil, ErrResourceNotServed
		}

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := h.cache.Get(name, namespace, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (h *resourceGetter) ListResources(gvr schema.GroupVersionResource, namespace string, query *query.Query) (client.ObjectList, error) {
	var obj client.ObjectList

	gvk, err := h.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	gvk = convertGVKToList(gvk)

	if h.client.Scheme().Recognizes(gvk) {
		gvkObject, err := h.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.ObjectList)
	} else {
		serviced, err := h.isServed(gvr)
		if err != nil {
			return nil, err
		}
		if !serviced {
			return nil, ErrResourceNotServed
		}
		u := &unstructured.UnstructuredList{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := h.cache.List(namespace, query, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func convertGVKToList(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	if strings.HasSuffix(gvk.Kind, "List") {
		return gvk
	}
	gvk.Kind = gvk.Kind + "List"
	return gvk
}

func (h *resourceGetter) getGVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	var (
		gvk schema.GroupVersionKind
		err error
	)
	gvk, err = h.client.RESTMapper().KindFor(gvr)
	if err != nil {
		return gvk, err
	}

	return gvk, nil
}

func (h *resourceGetter) isServed(gvr schema.GroupVersionResource) (bool, error) {
	resourceName := gvr.Resource + "." + gvr.Group
	h.RWMutex.RLock()
	isServed := h.serveCRD[resourceName]
	h.RWMutex.RUnlock()
	if isServed {
		return true, nil
	}

	crd := &extv1.CustomResourceDefinition{}
	err := h.client.Get(context.Background(), client.ObjectKey{Name: resourceName}, crd)
	if err != nil {
		return false, err
	}
	if crd.Labels[labelResourceServed] == "true" {
		h.RWMutex.Lock()
		h.serveCRD[resourceName] = true
		h.RWMutex.Unlock()
		return true, nil
	}
	return false, nil
}
