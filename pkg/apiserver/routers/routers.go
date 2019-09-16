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
	"fmt"
	"github.com/emicklei/go-restful"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"net/http"

	"kubesphere.io/kubesphere/pkg/server/errors"

	"strings"

	"k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/models/routers"
)

type Router struct {
	RouterType  string            `json:"type"`
	Annotations map[string]string `json:"annotations"`
}

// Get all namespace ingress controller services
func GetAllRouters(request *restful.Request, response *restful.Response) {

	routers, err := routers.GetAllRouters()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(routers)
}

// Get ingress controller service for specified namespace
func GetRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	router, err := routers.GetRouter(namespace)

	if err != nil {
		if k8serr.IsNotFound(err) {
			response.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		} else {
			response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	response.WriteAsJson(router)
}

// Create ingress controller and related services
func CreateRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := Router{}
	err := request.ReadEntity(&newRouter)

	if err != nil {
		response.WriteAsJson(err)
		return
	}

	var router *v1.Service

	serviceType, annotationMap, err := parseParameter(newRouter)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("wrong annotations, missing key or value")))
		return
	}

	router, err = routers.CreateRouter(namespace, serviceType, annotationMap)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(router)
}

// Delete ingress controller and services
func DeleteRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	router, err := routers.DeleteRouter(namespace)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(router)
}

func UpdateRouter(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	newRouter := Router{}
	err := request.ReadEntity(&newRouter)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	serviceType, annotationMap, err := parseParameter(newRouter)

	router, err := routers.UpdateRouter(namespace, serviceType, annotationMap)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(router)
}

func parseParameter(router Router) (routerType v1.ServiceType, annotationMap map[string]string, err error) {

	routerType = v1.ServiceTypeNodePort

	if strings.Compare(strings.ToLower(router.RouterType), "loadbalancer") == 0 {
		routerType = v1.ServiceTypeLoadBalancer
	}

	return routerType, router.Annotations, nil
}
