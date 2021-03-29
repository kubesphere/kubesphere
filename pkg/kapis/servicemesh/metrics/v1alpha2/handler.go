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
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
)

// default jaeger query api endpoint address
var JaegerQueryUrl = "http://jaeger-query.istio-system.svc:16686"

/*
Use Kiali API directly if config existed in configmap.
Such as:
kubectl -n kubesphere-system get cm kubesphere-config -oyaml
...
kialiQueryHost: http://kiali.istio-system:20001
...

Otherwise, use the API provided by kiali code.

Announce: The API provided by kiali code will deprecated in the future.
*/

var KialiQueryUrl string

// Get app metrics
func getAppMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	app := request.PathParameter("app")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/apps/%s/metrics?%s", KialiQueryUrl, namespace, app, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get workload metrics
func getWorkloadMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")

	if len(namespace) > 0 && len(workload) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s&workload=%s", request.Request.URL.RawQuery, namespace, workload)
	}

	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/workloads/%s/metrics?%s", KialiQueryUrl, namespace, workload, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get service metrics
func getServiceMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/services/%s/metrics?%s", KialiQueryUrl, namespace, service, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get namespace metrics
func getNamespaceMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/metrics?%s", KialiQueryUrl, namespace, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get service graph for namespace
func getNamespaceGraph(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	if len(namespace) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s", request.Request.URL.RawQuery, namespace)
	}

	url := fmt.Sprintf("%s/kiali/api/namespaces/graph?%s", KialiQueryUrl, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get namespace health
func getNamespaceHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/health?%s", KialiQueryUrl, namespace, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get workload health
func getWorkloadHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/workloads/%s/health?%s", KialiQueryUrl, namespace, workload, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get app health
func getAppHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	app := request.PathParameter("app")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/apps/%s/health?%s", KialiQueryUrl, namespace, app, request.Request.URL.RawQuery)
	getData(response, url)
}

// Get service health
func getServiceHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	url := fmt.Sprintf("%s/kiali/api/namespaces/%s/services/%s/health?%s", KialiQueryUrl, namespace, service, request.Request.URL.RawQuery)
	getData(response, url)
}

func getServiceTracing(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	serviceName := fmt.Sprintf("%s.%s", service, namespace)
	url := fmt.Sprintf("%s/api/traces?%s&service=%s", JaegerQueryUrl, request.Request.URL.RawQuery, serviceName)
	getData(response, url)
}

func getData(response *restful.Response, url string) {
	resp, err := http.Get(url)
	klog.V(4).Infof("Proxy request to %s", url)

	if err != nil {
		klog.Errorf("query url %s failed with err %v", url, err)
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
