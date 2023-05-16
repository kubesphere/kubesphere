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
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
)

var GroupVersion = schema.GroupVersion{Group: "gateway.kubesphere.io", Version: "v1alpha1"}

func AddToContainer(container *restful.Container, options *gateway.Options, cache cache.Cache, client client.Client, factory informers.InformerFactory, k8sClient kubernetes.Interface, loggingClient loggingclient.Client) error {
	ws := runtime.NewWebService(GroupVersion)

	handler := newHandler(options, cache, client, factory, k8sClient, loggingClient)

	// register gateway apis
	ws.Route(ws.POST("/namespaces/{namespace}/gateways").
		To(handler.Create).
		Doc("Create a gateway for a specified namespace.").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Param(ws.BodyParameter("gateway", "Gateway specification")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Reads(v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/gateways/").
		To(handler.Delete).
		Doc("Delete the specified gateway in namespace.").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/gateways/").
		To(handler.Update).
		Doc("Update gateway for a specified namespace.").
		Reads(v1alpha1.Gateway{}).
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Param(ws.BodyParameter("gateway", "Gateway specification")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/gateways/").
		To(handler.Get).
		Doc("Retrieve gateways details.").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.POST("/namespaces/{namespace}/gateways/{gateway}/upgrade").
		To(handler.Upgrade).
		Doc("Upgrade the legacy Project Gateway to the CRD based Gateway.").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Reads(v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.GET("/gateways/").
		To(handler.List).
		Doc("List Gateway details.").
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/gateways/{gateway}/pods").
		To(handler.ListPods).
		Doc("Retrieve gateways workload pods.").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/gateways/{gateway}/pods/{pod}/log").
		To(handler.PodLog).
		Doc("Retrieve log of the gateway's pod").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Param(ws.PathParameter("pod", "the pod name of the gateway")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Gateway{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/gateways/{gateway}/logs").
		To(handler.PodLogSearch).
		Doc("Retrieve log of the gateway's pod from ES").
		Param(ws.PathParameter("namespace", "the watching namespace of the gateway")).
		Param(ws.PathParameter("gateway", "the name of the gateway")).
		Returns(http.StatusOK, api.StatusOK, loggingv1alpha2.APIResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GatewayTag}))

	container.Add(ws)
	return nil
}
