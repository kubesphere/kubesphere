/*
Copyright 2018 The KubeSphere Authors.

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

package pods

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/filter/route"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

func Register(ws *restful.WebService) {

	ws.Route(ws.GET("/pods").To(handleAllPods).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(handlePodUnderNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/deployments/{deployment}/pods").
		To(handleGetDeploymentPodsMetrics).
		Filter(route.RouteLogging).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON))

	ws.Route(ws.GET("/namespaces/{namespace}/daemonsets/{daemonset}/pods").
		To(handleGetDaemonsetPodsMetrics).
		Filter(route.RouteLogging).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON))

	ws.Route(ws.GET("/namespaces/{namespace}/statefulsets/{statefulset}/pods").
		To(handleGetStatefulsetPodsMetrics).
		Filter(route.RouteLogging).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON))

	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(handlePodsUnderNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/nodes/{node}/pods").To(handlePodsUnderNode).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/nodes/{node}/namespaces/{namespace}/pods").To(handlePodsUnderNodeAndNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

// Get all pods metrics in cluster
func handleAllPods(_ *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	result = metrics.GetAllPodMetrics()
	response.WriteAsJson(result)
}

// Get pods metrics in namespace
func handlePodsUnderNameSpace(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	labelSelector := request.QueryParameter("labelSelector")
	result = metrics.GetPodMetricsInNamespace(request.PathParameter("namespace"), labelSelector)
	response.WriteAsJson(result)
}

// Get pods metrics in a deployment
func handleGetDeploymentPodsMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	deployment := request.PathParameter("deployment")
	result := metrics.GetPodMetricsInDeployment(namespace, deployment)
	response.WriteAsJson(result)
}

// Get pods metrics in daemonset deployment
func handleGetDaemonsetPodsMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	daemonset := request.PathParameter("daemonset")
	result := metrics.GetPodMetricsInDaemonset(namespace, daemonset)
	response.WriteAsJson(result)
}

// Get pods metrics in statefulset deployment
func handleGetStatefulsetPodsMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	statefulset := request.PathParameter("statefulset")
	result := metrics.GetPodMetricsInStatefulSet(namespace, statefulset)
	response.WriteAsJson(result)
}

// Get all pods metrics located in node
func handlePodsUnderNode(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	result = metrics.GetPodMetricsInNode(request.PathParameter("node"))
	response.WriteAsJson(result)
}

// Get a specific pod metrics
func handlePodUnderNameSpace(request *restful.Request, response *restful.Response) {
	var resultPod metrics.PodMetrics
	resultPod = metrics.FormatPodMetrics(request.PathParameter("namespace"), request.PathParameter("pod"))
	response.WriteAsJson(resultPod)
}

// Get pod metrics in a namespace located in deployment
func handlePodsUnderNodeAndNameSpace(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	nodeName := request.PathParameter("node")
	namespace := request.PathParameter("namespace")
	result = metrics.GetPodMetricsInNamespaceOfNode(namespace, nodeName)
	response.WriteAsJson(result)
}
