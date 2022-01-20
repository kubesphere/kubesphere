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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UnstructuredCRDHandler struct {
	c       client.Reader
	gvk     schema.GroupVersionKind
	gvkList schema.GroupVersionKind
}

func NewUnstructured(cache client.Reader, gvk schema.GroupVersionKind) Client {
	return &UnstructuredCRDHandler{c: cache, gvk: gvk, gvkList: schema.GroupVersionKind{Version: gvk.Version, Group: gvk.Group, Kind: gvk.Kind + "List"}}
}

func (h *UnstructuredCRDHandler) GetResources(request *restful.Request, response *restful.Response) {
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
func (h *UnstructuredCRDHandler) ListResources(request *restful.Request, response *restful.Response) {
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

func (h *UnstructuredCRDHandler) Get(key types.NamespacedName) (client.Object, error) {

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(h.gvk)

	if err := h.c.Get(context.TODO(), key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (h *UnstructuredCRDHandler) List(namespace string, q *query.Query) (*api.ListResult, error) {

	listObj := &unstructured.UnstructuredList{}
	listObj.SetGroupVersionKind(h.gvkList)

	if err := h.c.List(context.TODO(), listObj, &client.ListOptions{LabelSelector: q.Selector()}, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	result := DefaultList(listObj, q, h.compare, h.filter)

	return result, nil
}

func (d *UnstructuredCRDHandler) compare(left, right metav1.Object, field query.Field) bool {

	return DefaultObjectMetaCompare(left, right, field)
}

func (d *UnstructuredCRDHandler) filter(object metav1.Object, filter query.Filter) bool {
	//TODO Maybe we can use a json path Filter here
	return DefaultObjectMetaFilter(object, filter)
}
