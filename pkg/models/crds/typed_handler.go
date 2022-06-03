/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crds

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TypedCRDHandler struct {
	c   client.Reader
	s   *runtime.Scheme
	gvk schema.GroupVersionKind
}

func NewTyped(cache client.Reader, gvk schema.GroupVersionKind, s *runtime.Scheme) Client {
	return &TypedCRDHandler{c: cache, gvk: gvk, s: s}
}

func (h *TypedCRDHandler) GetResources(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	obj, err := h.Get(types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	})

	if err != nil {
		api.HandleError(response, nil, err)
		return
	}
	response.WriteEntity(obj)
}

// handleListResources retrieves resources
func (h *TypedCRDHandler) ListResources(request *restful.Request, response *restful.Response) {
	q := query.ParseQueryParameter(request)
	namespace := request.PathParameter("namespace")
	workspace := request.PathParameter("workspace")
	if workspace != "" {
		// filter by workspace
		q.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", "kubesphere.io/workspace", workspace))
	}
	list, err := h.List(namespace, q)
	if err != nil {
		api.HandleError(response, nil, err)
		return
	}
	response.WriteEntity(list)
}

func (h *TypedCRDHandler) Get(key types.NamespacedName) (client.Object, error) {
	obj, err := h.s.New(h.gvk)
	if err != nil {
		return nil, err
	}
	clobj := obj.(client.Object)
	if err := h.c.Get(context.TODO(), key, clobj); err != nil {
		return nil, err
	}
	return clobj, err
}

func (h *TypedCRDHandler) List(namespace string, q *query.Query) (*api.ListResult, error) {

	listGvk := schema.GroupVersionKind{
		Group:   h.gvk.Group,
		Version: h.gvk.Version,
		Kind:    h.gvk.Kind + "List",
	}
	obj, err := h.s.New(listGvk)
	if err != nil {
		return nil, err
	}
	objlist := obj.(client.ObjectList)
	if err := h.c.List(context.TODO(), objlist, &client.ListOptions{LabelSelector: q.Selector()}, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	result := DefaultList(objlist, q, h.compare, h.filter, h.transforms()...)

	return result, err
}

func (d *TypedCRDHandler) compare(left, right metav1.Object, field query.Field) bool {
	if fn, ok := Comparers[d.gvk]; ok {
		fn(left, right, field)
	}
	return DefaultObjectMetaCompare(left, right, field)
}

func (d *TypedCRDHandler) filter(object metav1.Object, filter query.Filter) bool {
	if fn, ok := Filters[d.gvk]; ok {
		fn(object, filter)
	}
	return DefaultObjectMetaFilter(object, filter)
}

func (d *TypedCRDHandler) transforms() []TransformFunc {
	if fn, ok := Transformers[d.gvk]; ok {
		return fn
	}
	trans := []TransformFunc{}
	return trans
}
