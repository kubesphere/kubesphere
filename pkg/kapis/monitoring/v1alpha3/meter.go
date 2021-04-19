package v1alpha3

import (
	"regexp"
	"strings"

	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/params"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func (h handler) HandleClusterMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) getAppWorkloads(ns string, apps []string) map[string][]string {
	return h.mo.GetAppWorkloads(ns, apps)
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
	appWorkloads := h.getAppWorkloads(aso.NamespaceName, aso.Applications)

	for k, _ := range appWorkloads {
		opt := monitoring.ApplicationOption{
			NamespaceName:         aso.NamespaceName,
			Application:           k,
			ApplicationComponents: appWorkloads[k],
			StorageClassName:      aso.StorageClassName,
		}

		if q.isRangeQuery() {
			current_res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, opt, h.meteringOptions.Billing.PriceInfo)
		} else {
			current_res, err = h.mo.GetNamedMeters(meters, q.time, opt, h.meteringOptions.Billing.PriceInfo)
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
		ExportMetrics(resp, res, q.start, q.end)
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
			current_res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, opt, h.meteringOptions.Billing.PriceInfo)
		} else {
			current_res, err = h.mo.GetNamedMeters(meters, q.time, opt, h.meteringOptions.Billing.PriceInfo)
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
		ExportMetrics(resp, res, q.start, q.end)
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

	_, ok = q.option.(monitoring.OpenpitrixsOption)
	if ok {
		h.handleOpenpitrixMetersQuery(meters, resp, q)
		return
	}

	_, ok = q.option.(monitoring.ServicesOption)
	if ok {
		h.handleServiceMetersQuery(meters, resp, q)
		return
	}

	if q.isRangeQuery() {
		res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, q.option, h.meteringOptions.Billing.PriceInfo)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
	} else {
		res, err = h.mo.GetNamedMeters(meters, q.time, q.option, h.meteringOptions.Billing.PriceInfo)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		if q.shouldSort() {
			res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		}
	}

	if q.Operation == OperationExport {
		ExportMetrics(resp, res, q.start, q.end)
		return
	}

	resp.WriteAsJson(res)
}

func (h handler) HandleNodeMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
	params.metering = true
	opt, err := h.makeQueryOptions(params, monitoring.LevelNode)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleWorkspaceMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
	params.metering = true
	opt, err := h.makeQueryOptions(params, monitoring.LevelWorkspace)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	h.handleNamedMetersQuery(resp, opt)
}

func (h handler) HandleNamespaceMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
	params.metering = true
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

func (h handler) HandleWorkloadMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) HandleApplicationMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) HandleOpenpitrixMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelOpenpitrix)
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

func (h handler) HandlePodMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) HandleServiceMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) HandlePVCMeterQuery(req *restful.Request, resp *restful.Response) {
	params := parseMeteringRequestParams(req)
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

func (h handler) collectOps(cluster, ns string) []string {

	var ops []string

	conditions := params.Conditions{
		Match: make(map[string]string),
		Fuzzy: make(map[string]string),
	}

	resp, err := h.opRelease.ListApplications("", cluster, ns, &conditions, 10, 0, "", false)
	if err != nil {
		klog.Error("failed to list op apps")
		return nil
	}
	totalCount := resp.TotalCount
	resp, err = h.opRelease.ListApplications("", cluster, ns, &conditions, totalCount, 0, "", false)
	if err != nil {
		klog.Error("failed to list op apps")
		return nil
	}

	for _, item := range resp.Items {
		app := item.(*openpitrix.Application)
		ops = append(ops, app.Cluster.ClusterId)
	}
	return ops
}
func (h handler) getOpWorkloads(cluster, ns string, ops []string) map[string][]string {

	componentsMap := make(map[string][]string)

	if len(ops) == 0 {
		ops = h.collectOps(cluster, ns)
	}

	for _, op := range ops {
		app, err := h.opRelease.DescribeApplication("", cluster, ns, op)
		if err != nil {
			klog.Error(err)
			return nil
		}
		for _, object := range app.ReleaseInfo {
			unstructuredObj := object.(*unstructured.Unstructured)
			componentsMap[op] = append(componentsMap[op], unstructuredObj.GetKind()+":"+unstructuredObj.GetName())
		}
	}

	return componentsMap
}
func (h handler) handleOpenpitrixMetersQuery(meters []string, resp *restful.Response, q queryOptions) {
	var metricMap = make(map[string]int)
	var res model.Metrics
	var current_res model.Metrics
	var err error

	oso, ok := q.option.(monitoring.OpenpitrixsOption)
	if !ok {
		klog.Error("invalid openpitrix option")
		return
	}

	opWorkloads := h.getOpWorkloads(oso.Cluster, oso.NamespaceName, oso.Openpitrixs)

	for k, _ := range opWorkloads {
		opt := monitoring.ApplicationOption{
			NamespaceName:         oso.NamespaceName,
			Application:           k,
			ApplicationComponents: opWorkloads[k],
			StorageClassName:      oso.StorageClassName,
		}

		if q.isRangeQuery() {
			current_res, err = h.mo.GetNamedMetersOverTime(meters, q.start, q.end, q.step, opt, h.meteringOptions.Billing.PriceInfo)
		} else {
			current_res, err = h.mo.GetNamedMeters(meters, q.time, opt, h.meteringOptions.Billing.PriceInfo)
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
		ExportMetrics(resp, res, q.start, q.end)
		return
	}

	resp.WriteAsJson(res)
}
