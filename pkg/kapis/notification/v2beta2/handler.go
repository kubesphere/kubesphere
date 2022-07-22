// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v2beta2

import (
	"io/ioutil"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	nmoperator "kubesphere.io/kubesphere/pkg/models/notification"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
)

type handler struct {
	operator nmoperator.Operator
}

func newNotificationHandler(
	informers informers.InformerFactory,
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface,
	options *notification.Options) *handler {

	return &handler{
		operator: nmoperator.NewOperator(informers, k8sClient, ksClient, options),
	}
}

func (h *handler) ListResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	subresource := req.QueryParameter("type")
	q := query.ParseQueryParameter(req)

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, subresource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s/%s", resource, subresource))
		return
	}

	objs, err := h.operator.List(user, resource, subresource, q)
	handleResponse(req, resp, objs, err)
}

func (h *handler) GetResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")
	subresource := req.QueryParameter("type")

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, subresource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s/%s", resource, subresource))
		return
	}

	obj, err := h.operator.Get(user, resource, name, subresource)
	handleResponse(req, resp, obj, err)
}

func (h *handler) CreateResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, "") {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	obj := h.operator.GetObject(resource, nmoperator.V2beta2)
	if err := req.ReadEntity(obj); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	created, err := h.operator.Create(user, resource, obj)
	handleResponse(req, resp, created, err)
}

func (h *handler) UpdateResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, "") {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	obj := h.operator.GetObject(resource, nmoperator.V2beta2)
	if err := req.ReadEntity(obj); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	updated, err := h.operator.Update(user, resource, name, obj)
	handleResponse(req, resp, updated, err)
}

func (h *handler) PatchResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, "") {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	data, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	patched, err := h.operator.Patch(user, resource, name, data)
	handleResponse(req, resp, patched, err)
}

func (h *handler) DeleteResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")

	if !h.operator.IsKnownResource(resource, nmoperator.V2beta2, "") {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	handleResponse(req, resp, servererr.None, h.operator.Delete(user, resource, name))
}

func (h *handler) Verify(req *restful.Request, resp *restful.Response) {
	h.operator.Verify(req, resp)
}

func handleResponse(req *restful.Request, resp *restful.Response, obj interface{}, err error) {

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		} else if errors.IsConflict(err) {
			api.HandleConflict(resp, req, err)
			return
		}
		api.HandleBadRequest(resp, req, err)
		return
	}

	_ = resp.WriteEntity(obj)
}
