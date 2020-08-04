/*
Copyright 2020 KubeSphere Authors

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

package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/kiali/kiali/handlers"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"net/http"
)

// default jaeger query api endpoint address
var JaegerQueryUrl = "http://jaeger-query.istio-system.svc:16686"

// Get app metrics
func getAppMetrics(request *restful.Request, response *restful.Response) {
	handlers.AppMetrics(request, response)
}

// Get workload metrics
func getWorkloadMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")

	if len(namespace) > 0 && len(workload) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s&workload=%s", request.Request.URL.RawQuery, namespace, workload)
	}

	handlers.WorkloadMetrics(request, response)
}

// Get service metrics
func getServiceMetrics(request *restful.Request, response *restful.Response) {
	handlers.ServiceMetrics(request, response)
}

// Get namespace metrics
func getNamespaceMetrics(request *restful.Request, response *restful.Response) {
	handlers.NamespaceMetrics(request, response)
}

// Get service graph for namespace
func getNamespaceGraph(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	if len(namespace) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s", request.Request.URL.RawQuery, namespace)
	}

	handlers.GetNamespaceGraph(request, response)
}

// Get service graph for namespaces
func getNamespacesGraph(request *restful.Request, response *restful.Response) {
	handlers.GraphNamespaces(request, response)
}

// Get namespace health
func getNamespaceHealth(request *restful.Request, response *restful.Response) {
	handlers.NamespaceHealth(request, response)
}

// Get workload health
func getWorkloadHealth(request *restful.Request, response *restful.Response) {
	handlers.WorkloadHealth(request, response)
}

// Get app health
func getAppHealth(request *restful.Request, response *restful.Response) {
	handlers.AppHealth(request, response)
}

// Get service health
func getServiceHealth(request *restful.Request, response *restful.Response) {
	handlers.ServiceHealth(request, response)
}

func getServiceTracing(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")

	serviceName := fmt.Sprintf("%s.%s", service, namespace)

	url := fmt.Sprintf("%s/api/traces?%s&service=%s", JaegerQueryUrl, request.Request.URL.RawQuery, serviceName)

	resp, err := http.Get(url)
	klog.V(4).Infof("Proxy trace request to %s", url)

	if err != nil {
		klog.Errorf("query jaeger failed with err %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		klog.Errorf("read response error : %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		klog.Errorf("write response failed %v", err)
	}
}
