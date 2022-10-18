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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/kiali"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
)

const (
	KubesphereNamespace      = "kubesphere-system"
	KubeSphereServiceAccount = "kubesphere"
)

type Handler struct {
	opt    *servicemesh.Options
	client *kiali.Client
}

func NewHandler(o *servicemesh.Options, client kubernetes.Interface, cache cache.Interface) *Handler {
	if o != nil && o.KialiQueryHost != "" {
		sa, err := client.CoreV1().ServiceAccounts(KubesphereNamespace).Get(context.TODO(), KubeSphereServiceAccount, metav1.GetOptions{})
		if err == nil {
			secret, err := client.CoreV1().Secrets(KubesphereNamespace).Get(context.TODO(), sa.Secrets[0].Name, metav1.GetOptions{})
			if err == nil {
				return &Handler{
					opt: o,
					client: kiali.NewDefaultClient(
						cache,
						string(secret.Data["token"]),
						o.KialiQueryHost,
					),
				}
			}
			klog.Warningf("get ServiceAccount's Secret failed %v", err)
		}
		klog.Warningf("get ServiceAccount failed %v", err)
	}
	// Handler should return Status code 400, instead of crash ks-apiserver
	// when no client is defined.
	return &Handler{opt: o, client: nil}
}

// Get app metrics
func (h *Handler) GetAppMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	app := request.PathParameter("app")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/apps/%s/metrics?%s", namespace, app, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get workload metrics
func (h *Handler) GetWorkloadMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")

	if len(namespace) > 0 && len(workload) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s&workload=%s", request.Request.URL.RawQuery, namespace, workload)
	}

	url := fmt.Sprintf("/kiali/api/namespaces/%s/workloads/%s/metrics?%s", namespace, workload, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get service metrics
func (h *Handler) GetServiceMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/services/%s/metrics?%s", namespace, service, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get namespace metrics
func (h *Handler) GetNamespaceMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/metrics?%s", namespace, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get service graph for namespace
func (h *Handler) GetNamespaceGraph(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	if len(namespace) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s", request.Request.URL.RawQuery, namespace)
	}

	url := fmt.Sprintf("/kiali/api/namespaces/graph?%s", request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get namespace health
func (h *Handler) GetNamespaceHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/health?%s", namespace, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get workload health
func (h *Handler) GetWorkloadHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/workloads/%s/health?%s", namespace, workload, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get app health
func (h *Handler) GetAppHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	app := request.PathParameter("app")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/apps/%s/health?%s", namespace, app, request.Request.URL.RawQuery)
	h.getData(response, url)
}

// Get service health
func (h *Handler) GetServiceHealth(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	url := fmt.Sprintf("/kiali/api/namespaces/%s/services/%s/health?%s", namespace, service, request.Request.URL.RawQuery)
	h.getData(response, url)
}

func (h *Handler) GetServiceTracing(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	serviceName := fmt.Sprintf("%s.%s", service, namespace)
	url := fmt.Sprintf("%s/api/traces?%s&service=%s", h.opt.JaegerQueryHost, request.Request.URL.RawQuery, serviceName)
	h.getJaegerData(response, url)
}

func (h *Handler) getData(response *restful.Response, url string) {

	if h.client == nil {
		err := errors.New("kiali url is not defined")
		api.HandleInternalError(response, nil, err)
		return
	}

	resp, err := h.client.Get(url)
	klog.V(4).Infof("Proxy request to %s", url)

	if err != nil {
		klog.Errorf("query url %s failed with err %v", url, err)
		api.HandleInternalError(response, nil, err)
		return
	}

	body, err := io.ReadAll(resp.Body)
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

// TODO: to be removed with a security Jaeger client
func (h *Handler) getJaegerData(response *restful.Response, url string) {

	resp, err := http.Get(url)
	klog.V(4).Infof("Proxy request to %s", url)

	if err != nil {
		klog.Errorf("query url %s failed with err %v", url, err)
		api.HandleInternalError(response, nil, err)
		return
	}

	body, err := io.ReadAll(resp.Body)
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
