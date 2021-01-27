/*
Copyright 2020 The KubeSphere Authors.

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

package v2

import (
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/notification"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
)

type handler struct {
	operator notification.Operator
}

func newNotificationHandler(
	informers informers.InformerFactory,
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface) *handler {

	return &handler{
		operator: notification.NewOperator(informers, k8sClient, ksClient),
	}
}

func (h *handler) ListSecret(req *restful.Request, resp *restful.Response) {
	q := query.ParseQueryParameter(req)
	objs, err := h.operator.ListSecret(q)
	handleResponse(req, resp, objs, err)
}

func (h *handler) GetSecret(req *restful.Request, resp *restful.Response) {

	obj, err := h.operator.GetSecret(req.PathParameter("secret"))
	handleResponse(req, resp, obj, err)
}

func (h *handler) CreateOrUpdateSecret(req *restful.Request, resp *restful.Response) {

	var obj corev1.Secret
	err := req.ReadEntity(&obj)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	created, err := h.operator.CreateOrUpdateSecret(&obj)
	handleResponse(req, resp, created, err)
}

func (h *handler) DeleteSecret(req *restful.Request, resp *restful.Response) {
	err := h.operator.DeleteSecret(req.PathParameter("secret"))
	handleResponse(req, resp, servererr.None, err)
}

func (h *handler) ListResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	q := query.ParseQueryParameter(req)

	if !h.operator.IsKnownResource(resource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	objs, err := h.operator.List(user, resource, q)
	handleResponse(req, resp, objs, err)
}

func (h *handler) GetResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")

	if !h.operator.IsKnownResource(resource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	obj, err := h.operator.Get(user, resource, name)
	handleResponse(req, resp, obj, err)
}

func (h *handler) CreateResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")

	if !h.operator.IsKnownResource(resource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	obj := h.operator.GetObject(resource)
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

	if !h.operator.IsKnownResource(resource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	obj := h.operator.GetObject(resource)
	if err := req.ReadEntity(obj); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	updated, err := h.operator.Update(user, resource, obj)
	handleResponse(req, resp, updated, err)
}

func (h *handler) DeleteResource(req *restful.Request, resp *restful.Response) {

	user := req.PathParameter("user")
	resource := req.PathParameter("resources")
	name := req.PathParameter("name")

	if !h.operator.IsKnownResource(resource) {
		api.HandleBadRequest(resp, req, servererr.New("unknown resource type %s", resource))
		return
	}

	handleResponse(req, resp, h.operator.Delete(user, resource, name), servererr.None)
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
