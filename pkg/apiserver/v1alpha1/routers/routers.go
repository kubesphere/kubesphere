/*

 Copyright 2019 The KubeSphere Authors.

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

package routers

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/errors"

	"net/http"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/models/routers"
)

func Route(ws *restful.WebService) {
	ws.Route(ws.GET("/routers").To(getAllRouters).
		Doc("Get all routers"))

	ws.Route(ws.GET("/users/{username}/routers").To(getAllRoutersOfUser).
		Doc("Get routers for user"))

	ws.Route(ws.GET("/namespaces/{namespace}/router").To(getRouter).
		Doc("Get router of a specified project").
		Param(ws.PathParameter("namespace", "name of the project").
			DataType("string")))

	ws.Route(ws.DELETE("/namespaces/{namespace}/router").To(deleteRouter).
		Doc("Get router of a specified project").
		Param(ws.PathParameter("namespace", "name of the project").
			DataType("string")))

	ws.Route(ws.POST("/namespaces/{namespace}/router").To(createRouter).
		Doc("Create a router for a specified project").
		Param(ws.PathParameter("namespace", "name of the project").
			DataType("string")))

	ws.Route(ws.PUT("/namespaces/{namespace}/router").To(updateRouter).
		Doc("Update a router for a specified project").
		Param(ws.PathParameter("namespace", "name of the project").
			DataType("string")))
}

type Router struct {
	RouterType  string            `json:"type"`
	Annotations map[string]string `json:"annotations"`
}

// Get all namespace ingress controller services
func getAllRouters(request *restful.Request, response *restful.Response) {

	routers, err := routers.GetAllRouters()

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(routers)
}

// Get all namespace ingress controller services for user
func getAllRoutersOfUser(request *restful.Request, response *restful.Response) {

	username := request.PathParameter("username")

	routers, err := routers.GetAllRoutersOfUser(username)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(routers)
}

// Get ingress controller service for specified namespace
func getRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	router, err := routers.GetRouter(namespace)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(router)
}

// Create ingress controller and related services
func createRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := Router{}
	err := request.ReadEntity(&newRouter)

	if err != nil {
		response.WriteAsJson(err)
		return
	}

	var router *v1.Service

	serviceType, annotationMap, err := ParseParameter(newRouter)

	if err != nil {
		glog.Error("Wrong annotations, missing key or value")
		response.WriteHeaderAndEntity(http.StatusBadRequest,
			errors.New(errors.InvalidArgument, "Wrong annotations, missing key or value"))
		return
	}

	router, err = routers.CreateRouter(namespace, serviceType, annotationMap)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(router)
}

// Delete ingress controller and services
func deleteRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	router, err := routers.DeleteRouter(namespace)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(router)
}

func updateRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := Router{}
	err := request.ReadEntity(&newRouter)

	if err != nil {
		glog.Error(err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	serviceType, annotationMap, err := ParseParameter(newRouter)

	router, err := routers.UpdateRouter(namespace, serviceType, annotationMap)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(router)
}

func ParseParameter(router Router) (routerType v1.ServiceType, annotationMap map[string]string, err error) {

	routerType = v1.ServiceTypeNodePort

	if strings.Compare(strings.ToLower(router.RouterType), "loadbalancer") == 0 {
		return v1.ServiceTypeLoadBalancer, router.Annotations, nil
	} else {
		return v1.ServiceTypeNodePort, make(map[string]string, 0), nil
	}

}
