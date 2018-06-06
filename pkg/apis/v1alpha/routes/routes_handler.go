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

package routes

import (
	"github.com/emicklei/go-restful"

	"errors"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService) {
	ws.Route(ws.GET("/routers").To(GetAllRouters).
		Doc("Get all routers").
		Filter(route.RouteLogging).
		Produces(restful.MIME_JSON))

	ws.Route(ws.GET("/{namespace}/router").To(GetRouter).
		Doc("Get router of a specified project").
		Param(ws.PathParameter("namespace", "name of the project").DataType("string")).
		Filter(route.RouteLogging).
		Produces(restful.MIME_JSON))

	ws.Route(ws.DELETE("/{namespace}/router").To(DeleteRouter).
		Doc("Get router of a specified project").
		Param(ws.PathParameter("namespace", "name of the project").DataType("string")).
		Filter(route.RouteLogging).
		Produces(restful.MIME_JSON))

	ws.Route(ws.POST("/{namespace}/router").To(CreateRouter).
		Doc("Create a router for a specified project").
		Param(ws.PathParameter("namespace", "name of the project").DataType("string")).
		Filter(route.RouteLogging).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON))

	ws.Route(ws.PUT("/{namespace}/router").To(UpdateRouter).
		Doc("Update a router for a specified project").
		Param(ws.PathParameter("namespace", "name of the project").DataType("string")).
		Filter(route.RouteLogging).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON))
}

// Get all namespace ingress controller services
func GetAllRouters(request *restful.Request, response *restful.Response) {

	routers, err := models.GetAllRouters()

	if err != nil {
		glog.Error(err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	} else {
		response.WriteAsJson(routers)
	}
}

// Get ingress controller service for specified namespace
func GetRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	router, err := models.GetRouter(namespace)

	if err != nil {
		glog.Error(err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	} else if router == nil {
		response.WriteHeaderAndEntity(http.StatusNotFound, constants.MessageResponse{Message: "Reseource Not Found"})
	} else {
		response.WriteAsJson(router)
	}

}

// Create ingress controller and related services
func CreateRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := models.Router{}
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
			constants.MessageResponse{Message: "Wrong annotations, missing key or value"})
		return
	}

	router, err = models.CreateRouter(namespace, serviceType, annotationMap)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	} else {
		response.WriteAsJson(*router)
	}
}

// Delete ingress controller and services
func DeleteRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	router, err := models.DeleteRouter(namespace)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	} else {
		response.WriteAsJson(router)
	}
}

func UpdateRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := models.Router{}
	err := request.ReadEntity(&newRouter)

	if err != nil {
		glog.Error(err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	serviceType, annotationMap, err := ParseParameter(newRouter)

	router, err := models.UpdateRouter(namespace, serviceType, annotationMap)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	} else {
		response.WriteAsJson(router)
	}
}

func ParseParameter(router models.Router) (routerType v1.ServiceType, annotationMap map[string]string, err error) {

	routerType = v1.ServiceTypeNodePort
	annotationMap = make(map[string]string)

	if strings.Compare(strings.ToLower(router.RouterType), "loadbalancer") == 0 {
		annotations := router.Annotations

		annotation := strings.FieldsFunc(annotations, func(r rune) bool {
			return r == ',' || r == '='
		})

		if len(annotation)%2 != 0 {
			glog.Error("Wrong annotations, missing key or value")
			return routerType, annotationMap, errors.New("wrong annotations, missing key or value")
		}

		for i := 0; i < len(annotation); i += 2 {
			annotationMap[annotation[i]] = annotation[i+1]
		}

		return v1.ServiceTypeLoadBalancer, annotationMap, nil
	} else {
		return v1.ServiceTypeNodePort, nil, nil
	}

}
