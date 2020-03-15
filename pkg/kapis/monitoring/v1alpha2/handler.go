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

package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type handler struct {
	k  k8s.Client
	mo model.MonitoringOperator
}

func newHandler(k k8s.Client, m monitoring.Interface) *handler {
	return &handler{k, model.NewMonitoringOperator(m)}
}

func (h handler) handleClusterMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelCluster)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleNodeMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelNode)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleWorkspaceMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelWorkspace)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleNamespaceMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelNamespace)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleWorkloadMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelWorkload)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handlePodMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelPod)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleContainerMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelContainer)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handlePVCMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelPVC)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleComponentMetricsQuery(req *restful.Request, resp *restful.Response) {
	p, err := h.parseRequestParams(req, monitoring.LevelComponent)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, p)
}

func (h handler) handleNamedMetricsQuery(resp *restful.Response, p params) {
	var res v1alpha2.APIResponse
	var err error

	if p.isRangeQuery() {
		res, err = h.mo.GetNamedMetricsOverTime(p.start, p.end, p.step, p.option)
		if err != nil {
			api.HandleInternalError(resp, nil, err)
			return
		}
	} else {
		res, err = h.mo.GetNamedMetrics(p.time, p.option)
		if err != nil {
			api.HandleInternalError(resp, nil, err)
			return
		}

		if p.shouldSort() {
			var rows int
			res, rows = h.mo.SortMetrics(res, p.target, p.order, p.identifier)
			res = h.mo.PageMetrics(res, p.page, p.limit, rows)
		}
	}

	resp.WriteAsJson(res)
}
