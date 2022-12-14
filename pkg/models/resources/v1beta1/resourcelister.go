package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"kubesphere.io/kubesphere/pkg/apiserver/query"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

func New(client client.Client, cache Interface) ResourceLister {
	return &resourceLister{
		client: client,
		cache:  cache,
	}
}

type resourceLister struct {
	client   client.Client
	cache    Interface
	gvrToGvk map[schema.GroupVersionResource]schema.GroupVersionKind
	sync.RWMutex
}

func (h *resourceLister) GetResource(gvr schema.GroupVersionResource, name, namespace string) (client.Object, error) {
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
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := h.cache.Get(name, namespace, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (h *resourceLister) ListResources(gvr schema.GroupVersionResource, namespace string, query *query.Query) (client.ObjectList, error) {
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

// TODO If can get the gvk of hot-plug crd?
func (h *resourceLister) getGVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	var (
		gvk schema.GroupVersionKind
		err error
	)
	gvk, err = h.client.RESTMapper().KindFor(gvr)
	if err == nil {
		return gvk, nil
	}

	if meta.IsNoMatchError(err) {
		gvk, err = h.getGVKByCRD(gvr)
		if err == nil {
			return gvk, nil
		}
	}
	return gvk, err
}

func (h *resourceLister) getGVKByCRD(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	var (
		gvk schema.GroupVersionKind
		ok  bool
	)
	h.RWMutex.RLock()
	gvk, ok = h.gvrToGvk[gvr]
	h.RWMutex.RUnlock()

	if ok {
		return gvk, nil
	}

	crd := extv1.CustomResourceDefinition{}
	err := h.client.Get(context.TODO(), client.ObjectKey{Name: fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group)}, &crd)
	if err != nil {
		return gvk, err
	}

	findVersion := false
	for _, v := range crd.Spec.Versions {
		if v.Name == gvr.Version && !v.Deprecated {
			findVersion = true
			break
		}
	}

	if !findVersion {
		return gvk, ErrResourceNotSupported
	}

	gvk.Group = gvr.Group
	gvk.Version = gvr.Version
	gvk.Kind = crd.Spec.Names.Kind

	h.RWMutex.Lock()
	h.gvrToGvk[gvr] = gvk
	h.RWMutex.Unlock()

	return gvk, nil
}
