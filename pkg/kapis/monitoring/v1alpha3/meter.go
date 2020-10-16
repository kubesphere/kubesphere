package v1alpha3

import (
	"regexp"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func (h handler) HandleClusterMetersQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelCluster)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetersQuery(resp, opt)
}

func getMetricPosMap(metrics []monitoring.Metric) map[string]int {
	var metricMap = make(map[string]int)

	for i, m := range metrics {
		metricMap[m.MetricName] = i
	}

	return metricMap
}

func (h handler) handleApplicationMetersQuery(meters []string, resp *restful.Response, q queryOptions) {
	var metricMap = make(map[string]int)
	var res model.Metrics
	var current_res model.Metrics
	var err error

	aso, ok := q.option.(monitoring.ApplicationsOption)
	if !ok {
		klog.Error("invalid application option")
		return
	}
	componentsMap := h.mo.GetAppComponentsMap(aso.NamespaceName, aso.Applications)

	for k, _ := range componentsMap {
		opt := monitoring.ApplicationOption{
			NamespaceName:         aso.NamespaceName,
			Application:           k,
			ApplicationComponents: componentsMap[k],
			StorageClassName:      aso.StorageClassName,
		}

		if q.isRangeQuery() {
			current_res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, opt)
		} else {
			current_res, err = h.mo.GetNamedMeters(meters, q.time, opt)
		}
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		if res.Results == nil {
			res = current_res
			metricMap = getMetricPosMap(res.Results)
		} else {
			for _, cur_res := range current_res.Results {
				pos, ok := metricMap[cur_res.MetricName]
				if ok {
					res.Results[pos].MetricValues = append(res.Results[pos].MetricValues, cur_res.MetricValues...)
				} else {
					res.Results = append(res.Results, cur_res)
				}
			}
		}
	}

	if !q.isRangeQuery() && q.shouldSort() {
		res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
	}

	if q.Operation == OperationExport {
		ExportMetrics(resp, res)
		return
	}

	resp.WriteAsJson(res)
}

func (h handler) handleServiceMetersQuery(meters []string, resp *restful.Response, q queryOptions) {
	var metricMap = make(map[string]int)
	var res model.Metrics
	var current_res model.Metrics
	var err error

	sso, ok := q.option.(monitoring.ServicesOption)
	if !ok {
		klog.Error("invalid service option")
		return
	}
	svcPodsMap := h.mo.GetSerivePodsMap(sso.NamespaceName, sso.Services)

	for k, _ := range svcPodsMap {
		opt := monitoring.ServiceOption{
			NamespaceName: sso.NamespaceName,
			ServiceName:   k,
			PodNames:      svcPodsMap[k],
		}

		if q.isRangeQuery() {
			current_res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, opt)
		} else {
			current_res, err = h.mo.GetNamedMeters(meters, q.time, opt)
		}
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		if res.Results == nil {
			res = current_res
			metricMap = getMetricPosMap(res.Results)
		} else {
			for _, cur_res := range current_res.Results {
				pos, ok := metricMap[cur_res.MetricName]
				if ok {
					res.Results[pos].MetricValues = append(res.Results[pos].MetricValues, cur_res.MetricValues...)
				} else {
					res.Results = append(res.Results, cur_res)
				}
			}
		}
	}

	if !q.isRangeQuery() && q.shouldSort() {
		res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
	}

	if q.Operation == OperationExport {
		ExportMetrics(resp, res)
		return
	}

	resp.WriteAsJson(res)
}

func (h handler) handleNamedMetersQuery(resp *restful.Response, q queryOptions) {
	var res model.Metrics
	var err error

	var meters []string
	for _, meter := range q.namedMetrics {
		if !strings.HasPrefix(meter, model.MetricMeterPrefix) {
			// skip non-meter metric
			continue
		}

		ok, _ := regexp.MatchString(q.metricFilter, meter)
		if ok {
			meters = append(meters, meter)
		}
	}

	if len(meters) == 0 {
		klog.Info("no meters found")
		resp.WriteAsJson(res)
		return
	}

	_, ok := q.option.(monitoring.ApplicationsOption)
	if ok {
		h.handleApplicationMetersQuery(meters, resp, q)
		return
	}

	_, ok = q.option.(monitoring.ServicesOption)
	if ok {
		h.handleServiceMetersQuery(meters, resp, q)
		return
	}

	if q.isRangeQuery() {
		res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, q.option)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
	} else {
		res, err = h.mo.GetNamedMeters(meters, q.time, q.option)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		if q.shouldSort() {
			res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		}
	}

	if q.Operation == OperationExport {
		ExportMetrics(resp, res)
		return
	}

	resp.WriteAsJson(res)
}

func (h handler) HandleNodeMetersQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelNode)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleWorkspaceMetersQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelWorkspace)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleNamespaceMetersQuery(req *restful.Request, resp *restful.Response) {
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

	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleWorkloadMetersQuery(req *restful.Request, resp *restful.Response) {
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
	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleApplicationMetersQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelApplication)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandlePodMetersQuery(req *restful.Request, resp *restful.Response) {
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
	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleServiceMetersQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelService)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}

	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandlePVCMetersQuery(req *restful.Request, resp *restful.Response) {
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
	h.handleNamedMetersQuery(resp, opt)
}
