package resource

import (
	"errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1/reader"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewResourceGetter(client client.Client,
	reader reader.Reader,
	t map[schema.GroupVersionResource]schema.GroupVersionKind) Interface {
	return &getter{
		client: client,
		reader: reader,
		t:      t,
	}
}

type getter struct {
	client client.Client
	reader reader.Reader
	t      map[schema.GroupVersionResource]schema.GroupVersionKind
}

func (g *getter) GetResource(gvr schema.GroupVersionResource, name, namespace string) (client.Object, error) {
	var obj client.Object
	if kindFor, err := g.client.RESTMapper().KindFor(gvr); err != nil {
		if meta.IsNoMatchError(err) {
			gvk, ok := g.t[gvr]
			if !ok {
				return nil, errors.New("not support type")
			}
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(gvk)
			obj = u
		} else {
			return nil, err
		}
	} else {
		gvkObject, err := g.client.Scheme().New(kindFor)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	}

	err := g.reader.Get(name, namespace, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (g *getter) ListResources(gvr schema.GroupVersionResource, query *query.Query) (client.ObjectList, error) {
	var objList client.ObjectList
	if kindFor, err := g.client.RESTMapper().KindFor(gvr); err != nil {
		if meta.IsNoMatchError(err) {
			gvk, ok := g.t[gvr]
			if !ok {
				return nil, errors.New("not support type")
			}
			u := &unstructured.UnstructuredList{}
			u.SetGroupVersionKind(gvk)
			objList = u
		} else {
			return nil, err
		}
	} else {
		gvkObject, err := g.client.Scheme().New(kindFor)
		if err != nil {
			return nil, err
		}
		objList = gvkObject.(client.ObjectList)
	}

	err := g.reader.List("", query, objList)
	if err != nil {
		return nil, err
	}
	return objList, nil
}
