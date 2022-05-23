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

package v1alpha3

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	converter "kubesphere.io/monitoring-dashboard/tools/converter"

	"kubesphere.io/kubesphere/pkg/models/openpitrix"

	"github.com/emicklei/go-restful"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	monitoringdashboardv1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/informers"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type handler struct {
	k               kubernetes.Interface
	mo              model.MonitoringOperator
	opRelease       openpitrix.ReleaseInterface
	meteringOptions *meteringclient.Options
	rtClient        runtimeclient.Client
}

func NewHandler(k kubernetes.Interface, monitoringClient monitoring.Interface, metricsClient monitoring.Interface, f informers.InformerFactory, resourceGetter *resourcev1alpha3.ResourceGetter, meteringOptions *meteringclient.Options, opClient openpitrix.Interface, rtClient runtimeclient.Client) *handler {

	if meteringOptions == nil || meteringOptions.RetentionDay == "" {
		meteringOptions = &meteringclient.DefaultMeteringOption
	}

	return &handler{
		k:               k,
		mo:              model.NewMonitoringOperator(monitoringClient, metricsClient, k, f, resourceGetter, opClient),
		opRelease:       opClient,
		meteringOptions: meteringOptions,
		rtClient:        rtClient,
	}
}

func (h handler) handleKubeSphereMetricsQuery(req *restful.Request, resp *restful.Response) {
	res := h.mo.GetKubeSphereStats()
	resp.WriteAsJson(res)
}

func (h handler) handleClusterMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelCluster)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleNodeMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelNode)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleWorkspaceMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelWorkspace)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	if req.QueryParameter("type") == "statistics" {
		res := h.mo.GetWorkspaceStats(params.workspaceName)
		resp.WriteAsJson(res)
	} else {
		h.handleNamedMetricsQuery(resp, opt)
	}
}

func (h handler) handleNamespaceMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelNamespace)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleWorkloadMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelWorkload)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handlePodMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelPod)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleContainerMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelContainer)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handlePVCMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelPVC)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleIngressMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelIngress)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleComponentMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func handleNoHit(namedMetrics []string) model.Metrics {
	var res model.Metrics
	for _, metic := range namedMetrics {
		res.Results = append(res.Results, monitoring.Metric{
			MetricName: metic,
			MetricData: monitoring.MetricData{},
		})
	}
	return res
}

func (h handler) handleNamedMetricsQuery(resp *restful.Response, q queryOptions) {
	var res model.Metrics

	var metrics []string
	for _, metric := range q.namedMetrics {
		if strings.HasPrefix(metric, model.MetricMeterPrefix) {
			// skip meter metric
			continue
		}
		ok, _ := regexp.MatchString(q.metricFilter, metric)
		if ok {
			metrics = append(metrics, metric)
		}
	}
	if len(metrics) == 0 {
		resp.WriteAsJson(res)
		return
	}

	if q.isRangeQuery() {
		res = h.mo.GetNamedMetricsOverTime(metrics, q.start, q.end, q.step, q.option)
	} else {
		res = h.mo.GetNamedMetrics(metrics, q.time, q.option)
		if q.shouldSort() {
			res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		}
	}
	resp.WriteAsJson(res)
}

func (h handler) handleMetadataQuery(req *restful.Request, resp *restful.Response) {
	res := h.mo.GetMetadata(req.PathParameter("namespace"))
	resp.WriteAsJson(res)
}

func (h handler) handleMetricLabelSetQuery(req *restful.Request, resp *restful.Response) {
	var res model.MetricLabelSet

	params := parseRequestParams(req)
	if params.metric == "" || params.start == "" || params.end == "" {
		api.HandleBadRequest(resp, nil, errors.New("required fields are missing: [metric, start, end]"))
		return
	}

	opt, err := h.makeQueryOptions(params, 0)
	if err != nil {
		if err.Error() == ErrNoHit {
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}

	res = h.mo.GetMetricLabelSet(params.metric, params.namespaceName, opt.start, opt.end)
	resp.WriteAsJson(res)
}

func (h handler) handleAdhocQuery(req *restful.Request, resp *restful.Response) {
	var res monitoring.Metric

	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, 0)
	if err != nil {
		if err.Error() == ErrNoHit {
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}

	if opt.isRangeQuery() {
		res, err = h.mo.GetMetricOverTime(params.expression, params.namespaceName, opt.start, opt.end, opt.step)
	} else {
		res, err = h.mo.GetMetric(params.expression, params.namespaceName, opt.time)
	}

	if err != nil {
		api.HandleBadRequest(resp, nil, err)
	} else {
		resp.WriteAsJson(res)
	}
}

// handleGrafanaDashboardImport imports Grafana template and converts it to KubeSphere dashboard.
// The description of the Parameters:
// grafanaDashboardName: the name of this Grafana template needed to convert.
// grafanaDashboardUrl: the link to download this Grafana template.
// grafanaDashboardContent: the whole JSON content needed to convert.
// Note that the parameter grafanaDashboardName is indispensable,
// and the requested parameter grafanaDashboardUrl and grafanaDashboardContent cannot be empty at the same time.
func (h handler) handleGrafanaDashboardImport(req *restful.Request, resp *restful.Response) {
	var entity monitoring.DashboardEntity
	err := req.ReadEntity(&entity)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	grafanaDashboardName := req.PathParameter("grafanaDashboardName")
	namespace := req.PathParameter("namespace")

	if grafanaDashboardName == "" {
		err := errors.New("the requested parameter grafanaDashboardName cannot be empty")
		api.HandleBadRequest(resp, nil, err)
		return
	}
	if entity.GrafanaDashboardUrl == "" && entity.GrafanaDashboardContent == "" {
		err := errors.New("the requested parameter grafanaDashboardUrl and grafanaDashboardContent cannot be empty at the same time")
		api.HandleBadRequest(resp, nil, err)
		return
	}

	// download the Grafana dashboard
	grafanaDashboardContent := []byte(entity.GrafanaDashboardContent)
	if entity.GrafanaDashboardUrl != "" {
		c, err := func(u string) ([]byte, error) {
			_, err := url.ParseRequestURI(u)
			if err != nil {
				return nil, err
			}
			client := &http.Client{}
			req, err := http.NewRequest("GET", u, nil)
			if err != nil {
				return nil, err
			}

			r, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			defer r.Body.Close()

			c, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			return c, nil
		}(entity.GrafanaDashboardUrl)

		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		grafanaDashboardContent = []byte(c)
	}

	isClusterCrd := namespace == ""
	c := converter.NewConverter()
	convertedDashboard, err := c.ConvertToDashboard(grafanaDashboardContent, isClusterCrd, namespace, grafanaDashboardName)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	ctx := context.TODO()
	annotation := map[string]string{"kubesphere.io/description": entity.Description}

	// a cluster scope dashboard or a namespaced dashboard with the same name cannot post.
	if isClusterCrd {
		clusterdashboard := monitoringdashboardv1alpha2.ClusterDashboard{
			TypeMeta: v1.TypeMeta{
				APIVersion: convertedDashboard.APIVersion,
				Kind:       convertedDashboard.Kind,
			},
			ObjectMeta: v1.ObjectMeta{
				Name:        convertedDashboard.Metadata["name"],
				Annotations: annotation,
			},
			Spec: *convertedDashboard.Spec,
		}

		objKey := runtimeclient.ObjectKey{
			Namespace: "",
			Name:      clusterdashboard.Name,
		}

		err = h.rtClient.Get(ctx, objKey, &clusterdashboard)

		if err == nil {
			api.HandleBadRequest(resp, nil, errors.New("dashboards with the same name already exists."))
			return
		}

		// create this dashboard
		err = h.rtClient.Create(ctx, &clusterdashboard)

		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		resp.WriteAsJson(clusterdashboard)

	} else {
		dashboard := monitoringdashboardv1alpha2.Dashboard{
			TypeMeta: v1.TypeMeta{
				APIVersion: convertedDashboard.APIVersion,
				Kind:       convertedDashboard.Kind,
			},
			ObjectMeta: v1.ObjectMeta{
				Name:        convertedDashboard.Metadata["name"],
				Namespace:   namespace,
				Annotations: annotation,
			},
			Spec: *convertedDashboard.Spec,
		}

		objKey := runtimeclient.ObjectKey{
			Namespace: namespace,
			Name:      dashboard.Name,
		}

		err = h.rtClient.Get(ctx, objKey, &dashboard)

		if err == nil {
			api.HandleBadRequest(resp, nil, errors.New("dashboards with the same name already exists."))
			return
		}

		// create this dashboard
		err = h.rtClient.Create(ctx, &dashboard)

		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		resp.WriteAsJson(dashboard)

	}

}
