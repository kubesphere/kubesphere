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
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"regexp"
)

type handler struct {
	k  kubernetes.Interface
	mo model.MonitoringOperator
}

func newHandler(k kubernetes.Interface, m monitoring.Interface) *handler {
	return &handler{k, model.NewMonitoringOperator(m)}
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
	h.handleNamedMetricsQuery(resp, opt)
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
