/*
Copyright 2021 KubeSphere Authors

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

package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	operator "kubesphere.io/kubesphere/pkg/models/gateway"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
)

type handler struct {
	options *gateway.Options
	gw      operator.GatewayOperator
}

//newHandler create an instance of the handler
func newHandler(options *gateway.Options, client client.Client) *handler {
	// Do not register Gateway scheme globally. Which will cause conflict in ks-controller-manager.
	v1alpha1.AddToScheme(client.Scheme())
	return &handler{
		options: options,
		gw:      operator.NewGatewayOperator(client, options),
	}
}

func (h *handler) Create(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")
	var gateway v1alpha1.Gateway

	err := request.ReadEntity(&gateway)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.gw.CreateGateway(ns, &gateway)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *handler) Update(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")
	var gateway v1alpha1.Gateway
	err := request.ReadEntity(&gateway)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.gw.UpdateGateway(ns, &gateway)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) Get(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")
	gateway, err := h.gw.GetGateways(ns)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(gateway)
}

func (h *handler) Delete(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")

	err := h.gw.DeleteGateway(ns)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *handler) Upgrade(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")

	g, err := h.gw.UpgradeGateway(ns)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(g)
}

func (h *handler) List(request *restful.Request, response *restful.Response) {
	//TODO
}
