package v1beta1

import (
	"errors"
	"strings"

	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

func New(client client.Client, cache Interface) ResourceGetter {
	return &resourceGetter{
		client: client,
		cache:  cache,
	}
}

type resourceGetter struct {
	client client.Client
	cache  Interface
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
