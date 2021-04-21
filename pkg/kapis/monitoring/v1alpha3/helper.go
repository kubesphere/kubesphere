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

package v1alpha3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"kubesphere.io/kubesphere/pkg/api"

	"github.com/jszwec/csvutil"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	DefaultStep   = 10 * time.Minute
	DefaultFilter = ".*"
	DefaultOrder  = model.OrderDescending
	DefaultPage   = 1
	DefaultLimit  = 5

	OperationQuery  = "query"
	OperationExport = "export"

	ComponentEtcd      = "etcd"
	ComponentAPIServer = "apiserver"
	ComponentScheduler = "scheduler"

	ErrNoHit             = "'end' or 'time' must be after the namespace creation time."
	ErrParamConflict     = "'time' and the combination of 'start' and 'end' are mutually exclusive."
	ErrInvalidStartEnd   = "'start' must be before 'end'."
	ErrInvalidPage       = "Invalid parameter 'page'."
	ErrInvalidLimit      = "Invalid parameter 'limit'."
	ErrParameterNotfound = "Parmameter [%s] not found"
)

type reqParams struct {
	metering                  bool
	operation                 string
	time                      string
	start                     string
	end                       string
	step                      string
	target                    string
	order                     string
	page                      string
	limit                     string
	metricFilter              string
	namespacedResourcesFilter string
	resourceFilter            string
	nodeName                  string
	workspaceName             string
	namespaceName             string
	workloadKind              string
	workloadName              string
	podName                   string
	containerName             string
	pvcName                   string
	storageClassName          string
	componentType             string
	expression                string
	metric                    string
	applications              string
	openpitrixs               string
	cluster                   string
	services                  string
	pvcFilter                 string
	queryType                 string
}

type queryOptions struct {
	metricFilter string
	namedMetrics []string

	Operation string

	start time.Time
	end   time.Time
	time  time.Time
	step  time.Duration

	target     string
	identifier string
	order      string
	page       int
	limit      int

	option monitoring.QueryOption
}

func (q queryOptions) isRangeQuery() bool {
	return q.time.IsZero()
}

func (q queryOptions) shouldSort() bool {
	return q.target != "" && q.identifier != ""
}

func parseRequestParams(req *restful.Request) reqParams {
	var r reqParams
	r.time = req.QueryParameter("time")
	r.start = req.QueryParameter("start")
	r.end = req.QueryParameter("end")
	r.step = req.QueryParameter("step")
	r.target = req.QueryParameter("sort_metric")
	r.order = req.QueryParameter("sort_type")
	r.page = req.QueryParameter("page")
	r.limit = req.QueryParameter("limit")
	r.metricFilter = req.QueryParameter("metrics_filter")
	// namespacedResourcesFilter supports only <namespace>/<pod_name>|<namespace>/<pod_name> format
	// which is different from resources_filter or metrics_filter, so wipe off the possible $ at the end.
	r.namespacedResourcesFilter = strings.TrimRight(req.QueryParameter("namespaced_resources_filter"), "$")
	r.resourceFilter = req.QueryParameter("resources_filter")
	r.workspaceName = req.PathParameter("workspace")
	r.namespaceName = req.PathParameter("namespace")
	r.workloadKind = req.PathParameter("kind")
	r.nodeName = req.PathParameter("node")
	r.workloadName = req.PathParameter("workload")
	r.podName = req.PathParameter("pod")
	r.containerName = req.PathParameter("container")
	r.pvcName = req.PathParameter("pvc")
	r.storageClassName = req.PathParameter("storageclass")
	r.componentType = req.PathParameter("component")
	r.expression = req.QueryParameter("expr")
	r.metric = req.QueryParameter("metric")
	r.queryType = req.QueryParameter("type")

	return r
}

func parseMeteringRequestParams(req *restful.Request) reqParams {
	params := parseRequestParams(req)

	// mark this request is metering req
	params.metering = true

	// whether need to export metering data
	params.operation = req.QueryParameter("operation")

	// OpenPitrix belongs to which cluster
	params.cluster = req.PathParameter("cluster")

	// specified which application crds
	params.applications = req.QueryParameter("applications")

	// specified which OpenPitrix apps
	params.openpitrixs = req.QueryParameter("openpitrix_ids")

	// specified which service
	params.services = req.QueryParameter("services")

	// specified which pvc
	params.pvcFilter = req.QueryParameter("pvc_filter")

	// support node param in URL query
	if req.QueryParameter("node") != "" {
		params.nodeName = req.QueryParameter("node")
	}

	// support kind param in URL query
	if req.QueryParameter("kind") != "" {
		params.workloadKind = req.QueryParameter("kind")
	}

	return params
}

func (h handler) makeQueryOptions(r reqParams, lvl monitoring.Level) (q queryOptions, err error) {
	if r.resourceFilter == "" {
		r.resourceFilter = DefaultFilter
	}

	q.metricFilter = r.metricFilter
	if r.metricFilter == "" {
		q.metricFilter = DefaultFilter
	}

	q.Operation = r.operation
	if r.operation == "" {
		q.Operation = OperationQuery
	}

	switch lvl {
	case monitoring.LevelCluster:
		q.option = monitoring.ClusterOption{}
		q.namedMetrics = model.ClusterMetrics

	case monitoring.LevelNode:
		q.identifier = model.IdentifierNode
		q.option = monitoring.NodeOption{
			ResourceFilter:   r.resourceFilter,
			NodeName:         r.nodeName,
			PVCFilter:        r.pvcFilter,        // metering pvc
			StorageClassName: r.storageClassName, // metering pvc
			QueryType:        r.queryType,
		}
		q.namedMetrics = model.NodeMetrics

	case monitoring.LevelWorkspace:
		q.identifier = model.IdentifierWorkspace
		q.option = monitoring.WorkspaceOption{
			ResourceFilter:   r.resourceFilter,
			WorkspaceName:    r.workspaceName,
			PVCFilter:        r.pvcFilter,        // metering pvc
			StorageClassName: r.storageClassName, // metering pvc
		}
		q.namedMetrics = model.WorkspaceMetrics

	case monitoring.LevelNamespace:
		q.identifier = model.IdentifierNamespace
		q.option = monitoring.NamespaceOption{
			ResourceFilter:   r.resourceFilter,
			WorkspaceName:    r.workspaceName,
			NamespaceName:    r.namespaceName,
			PVCFilter:        r.pvcFilter,        // metering pvc
			StorageClassName: r.storageClassName, // metering pvc
		}
		q.namedMetrics = model.NamespaceMetrics

	case monitoring.LevelApplication:
		q.identifier = model.IdentifierApplication
		if r.namespaceName == "" {
			return q, errors.New(fmt.Sprintf(ErrParameterNotfound, "namespace"))
		}

		application := []string{}
		if len(r.applications) != 0 {
			application = strings.Split(r.applications, "|")
		}

		q.option = monitoring.ApplicationsOption{
			NamespaceName:    r.namespaceName,
			Applications:     application,
			StorageClassName: r.storageClassName, // metering pvc
		}
		q.namedMetrics = model.ApplicationMetrics

	case monitoring.LevelOpenpitrix:
		q.identifier = model.IdentifierApplication
		if r.namespaceName == "" {
			return q, errors.New(fmt.Sprintf(ErrParameterNotfound, "namespace"))
		}

		ops := []string{}
		if len(r.openpitrixs) != 0 {
			ops = strings.Split(r.openpitrixs, "|")
		}
		q.option = monitoring.OpenpitrixsOption{
			Cluster:          r.cluster,
			NamespaceName:    r.namespaceName,
			Openpitrixs:      ops,
			StorageClassName: r.storageClassName,
		}

		// op share the same metrics with application
		q.namedMetrics = model.ApplicationMetrics

	case monitoring.LevelWorkload:
		q.identifier = model.IdentifierWorkload
		q.option = monitoring.WorkloadOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
		}
		q.namedMetrics = model.WorkloadMetrics

	case monitoring.LevelPod:
		q.identifier = model.IdentifierPod
		q.option = monitoring.PodOption{
			NamespacedResourcesFilter: r.namespacedResourcesFilter,
			ResourceFilter:            r.resourceFilter,
			NodeName:                  r.nodeName,
			NamespaceName:             r.namespaceName,
			WorkloadKind:              r.workloadKind,
			WorkloadName:              r.workloadName,
			PodName:                   r.podName,
		}
		q.namedMetrics = model.PodMetrics

	case monitoring.LevelService:
		q.identifier = model.IdentifierService

		svcs := []string{}
		if len(r.services) != 0 {
			svcs = strings.Split(r.services, "|")
		}

		q.option = monitoring.ServicesOption{
			NamespaceName: r.namespaceName,
			Services:      svcs,
		}

		q.namedMetrics = model.ServiceMetrics

	case monitoring.LevelContainer:
		q.identifier = model.IdentifierContainer
		q.option = monitoring.ContainerOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			PodName:        r.podName,
			ContainerName:  r.containerName,
		}
		q.namedMetrics = model.ContainerMetrics

	case monitoring.LevelPVC:
		q.identifier = model.IdentifierPVC
		q.option = monitoring.PVCOption{
			ResourceFilter:            r.resourceFilter,
			NamespaceName:             r.namespaceName,
			StorageClassName:          r.storageClassName,
			PersistentVolumeClaimName: r.pvcName,
		}
		q.namedMetrics = model.PVCMetrics

	case monitoring.LevelComponent:
		q.option = monitoring.ComponentOption{}
		switch r.componentType {
		case ComponentEtcd:
			q.namedMetrics = model.EtcdMetrics
		case ComponentAPIServer:
			q.namedMetrics = model.APIServerMetrics
		case ComponentScheduler:
			q.namedMetrics = model.SchedulerMetrics
		}
	}

	// Parse time params
	if r.start != "" && r.end != "" {
		startInt, err := strconv.ParseInt(r.start, 10, 64)
		if err != nil {
			return q, err
		}
		q.start = time.Unix(startInt, 0)

		endInt, err := strconv.ParseInt(r.end, 10, 64)
		if err != nil {
			return q, err
		}
		q.end = time.Unix(endInt, 0)

		if r.step == "" {
			q.step = DefaultStep
		} else {
			q.step, err = time.ParseDuration(r.step)
			if err != nil {
				return q, err
			}
		}

		if q.start.After(q.end) {
			return q, errors.New(ErrInvalidStartEnd)
		}
	} else if r.start == "" && r.end == "" {
		if r.time == "" {
			q.time = time.Now()
		} else {
			timeInt, err := strconv.ParseInt(r.time, 10, 64)
			if err != nil {
				return q, err
			}
			q.time = time.Unix(timeInt, 0)
		}
	} else {
		return q, errors.Errorf(ErrParamConflict)
	}

	// Ensure query start time to be after the namespace creation time
	if r.namespaceName != "" && !r.metering {
		ns, err := h.k.CoreV1().Namespaces().Get(context.Background(), r.namespaceName, corev1.GetOptions{})
		if err != nil {
			return q, err
		}
		cts := ns.CreationTimestamp.Time

		// Query should happen no earlier than namespace's creation time.
		// For range query, check and mutate `start`. For instant query, check `time`.
		// In range query, if `start` and `end` are both before namespace's creation time, it causes no hit.
		if !q.isRangeQuery() {
			if q.time.Before(cts) {
				return q, errors.New(ErrNoHit)
			}
		} else {
			if q.end.Before(cts) {
				return q, errors.New(ErrNoHit)
			}
			if q.start.Before(cts) {
				q.start = q.end
				for q.start.Add(-q.step).After(cts) {
					q.start = q.start.Add(-q.step)
				}
			}
		}
	}

	// Parse sorting and paging params
	if r.target != "" {
		q.target = r.target
		q.page = DefaultPage
		q.limit = DefaultLimit
		q.order = r.order
		if r.order != model.OrderAscending {
			q.order = DefaultOrder
		}
		if r.page != "" {
			q.page, err = strconv.Atoi(r.page)
			if err != nil || q.page <= 0 {
				return q, errors.New(ErrInvalidPage)
			}
		}
		if r.limit != "" {
			q.limit, err = strconv.Atoi(r.limit)
			if err != nil || q.limit <= 0 {
				return q, errors.New(ErrInvalidLimit)
			}
		}
	}

	return q, nil
}

func exportMetrics(metrics model.Metrics, startTime, endTime time.Time) (*bytes.Buffer, error) {
	var resBytes []byte

	for i, _ := range metrics.Results {
		ret := metrics.Results[i]
		for j, _ := range ret.MetricValues {
			ret.MetricValues[j].TransferToExportedMetricValue()
		}
	}

	for _, metric := range metrics.Results {

		metricName := metric.MetricName

		var csvpoints []monitoring.CSVPoint
		for _, metricVal := range metric.MetricValues {

			var targetList []string
			for k, v := range metricVal.Metadata {
				targetList = append(targetList, fmt.Sprintf("%s=%s", k, v))
			}
			selector := strings.Join(targetList, "|")

			statsTab := "\nmetric_name,selector,start_time,end_time,min,max,avg,sum,fee, currency_unit\n" +
				fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n\n",
					metricName,
					selector,
					startTime.String(),
					endTime.String(),
					metricVal.MinValue,
					metricVal.MaxValue,
					metricVal.AvgValue,
					metricVal.SumValue,
					metricVal.Fee,
					metricVal.CurrencyUnit)

			csvpoints = nil
			resourceUnit := metricVal.ResourceUnit
			for _, p := range metricVal.ExportedSeries {
				csvpoints = append(csvpoints, p.TransformToCSVPoint(metricName, selector, resourceUnit))
			}

			dataTab, err := csvutil.Marshal(csvpoints)
			if err != nil {
				return nil, err
			}

			resBytes = append(resBytes, statsTab...)
			resBytes = append(resBytes, dataTab...)
		}
	}

	if len(resBytes) == 0 {
		resBytes = []byte("no data")
	}

	output := new(bytes.Buffer)
	_, err := output.Write(resBytes)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func ExportMetrics(resp *restful.Response, metrics model.Metrics, startTime, endTime time.Time) {
	resp.Header().Set(restful.HEADER_ContentType, "text/plain")
	resp.Header().Set("Content-Disposition", "attachment")

	output, err := exportMetrics(metrics, startTime, endTime)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	_, err = io.Copy(resp, output)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	return
}
