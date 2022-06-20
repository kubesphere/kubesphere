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
	"fmt"
	"time"

	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/util/flushwriter"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/api/gateway/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	operator "kubesphere.io/kubesphere/pkg/models/gateway"
	"kubesphere.io/kubesphere/pkg/models/logging"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	conversionsv1 "kubesphere.io/kubesphere/pkg/utils/conversions/core/v1"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

type handler struct {
	options *gateway.Options
	gw      operator.GatewayOperator
	factory informers.InformerFactory
	lo      logging.LoggingOperator
}

//newHandler create an instance of the handler
func newHandler(options *gateway.Options, cache cache.Cache, client client.Client, factory informers.InformerFactory, k8sClient kubernetes.Interface, loggingClient loggingclient.Client) *handler {
	conversionsv1.RegisterConversions(scheme.Scheme)
	// Do not register Gateway scheme globally. Which will cause conflict in ks-controller-manager.
	v1alpha1.AddToScheme(client.Scheme())
	var lo logging.LoggingOperator
	if loggingClient != nil {
		lo = logging.NewLoggingOperator(loggingClient)
	}
	return &handler{
		options: options,
		factory: factory,
		gw:      operator.NewGatewayOperator(client, cache, options, factory, k8sClient),
		lo:      lo,
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
	queryParam := query.ParseQueryParameter(request)

	result, err := h.gw.ListGateways(queryParam)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *handler) ListPods(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	ns := request.PathParameter("namespace")

	result, err := h.gw.GetPods(ns, queryParam)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *handler) PodLog(request *restful.Request, response *restful.Response) {

	podNamespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")

	query := request.Request.URL.Query()
	logOptions := &corev1.PodLogOptions{}
	if err := scheme.ParameterCodec.DecodeParameters(query, corev1.SchemeGroupVersion, logOptions); err != nil {
		api.HandleError(response, request, fmt.Errorf("unable to decode query"))
		return
	}

	fw := flushwriter.Wrap(response.ResponseWriter)
	err := h.gw.GetPodLogs(request.Request.Context(), podNamespace, podID, logOptions, fw)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
}

func (h *handler) PodLogSearch(request *restful.Request, response *restful.Response) {
	if h.lo == nil {
		api.HandleError(response, request, fmt.Errorf("logging isn't enabled"))
		return
	}

	ns := request.PathParameter("namespace")
	logQuery, err := loggingv1alpha2.ParseQueryParameter(request)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	// ES log will be filtered by pods and namespace by default.
	pods, err := h.gw.GetPods(ns, &query.Query{})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var podfilter []string
	namespaceCreateTimeMap := make(map[string]*time.Time)
	var ar loggingv1alpha2.APIResponse

	for _, p := range pods.Items {
		pod, ok := p.(*corev1.Pod)
		if ok {
			podfilter = append(podfilter, pod.Name)
			namespaceCreateTimeMap[pod.Namespace] = nil
		}
	}

	sf := loggingclient.SearchFilter{
		NamespaceFilter: namespaceCreateTimeMap,
		PodFilter:       podfilter,
		PodSearch:       stringutils.Split(logQuery.PodSearch, ","),
		ContainerSearch: stringutils.Split(logQuery.ContainerSearch, ","),
		ContainerFilter: stringutils.Split(logQuery.ContainerFilter, ","),
		LogSearch:       stringutils.Split(logQuery.LogSearch, ","),
		Starttime:       logQuery.StartTime,
		Endtime:         logQuery.EndTime,
	}

	noHit := len(namespaceCreateTimeMap) == 0 || len(podfilter) == 0

	if logQuery.Operation == loggingv1alpha2.OperationExport {
		response.Header().Set(restful.HEADER_ContentType, "text/plain")
		response.Header().Set("Content-Disposition", "attachment")
		if noHit {
			return
		}

		err = h.lo.ExportLogs(sf, response)
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
	} else {
		if noHit {
			ar.Logs = &loggingclient.Logs{}
		}

		ar, err = h.lo.SearchLogs(sf, logQuery.From, logQuery.Size, logQuery.Sort)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		response.WriteEntity(ar)
	}

}
