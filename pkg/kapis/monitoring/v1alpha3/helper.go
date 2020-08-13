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
	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"strconv"
	"time"
)

const (
	DefaultStep   = 10 * time.Minute
	DefaultFilter = ".*"
	DefaultOrder  = model.OrderDescending
	DefaultPage   = 1
	DefaultLimit  = 5

	ComponentEtcd      = "etcd"
	ComponentAPIServer = "apiserver"
	ComponentScheduler = "scheduler"

	ErrNoHit           = "'end' or 'time' must be after the namespace creation time."
	ErrParamConflict   = "'time' and the combination of 'start' and 'end' are mutually exclusive."
	ErrInvalidStartEnd = "'start' must be before 'end'."
	ErrInvalidPage     = "Invalid parameter 'page'."
	ErrInvalidLimit    = "Invalid parameter 'limit'."
)

type reqParams struct {
	time             string
	start            string
	end              string
	step             string
	target           string
	order            string
	page             string
	limit            string
	metricFilter     string
	resourceFilter   string
	nodeName         string
	workspaceName    string
	namespaceName    string
	workloadKind     string
	workloadName     string
	podName          string
	containerName    string
	pvcName          string
	storageClassName string
	componentType    string
	expression       string
	metric           string
}

type queryOptions struct {
	metricFilter string
	namedMetrics []string

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
	r.resourceFilter = req.QueryParameter("resources_filter")
	r.nodeName = req.PathParameter("node")
	r.workspaceName = req.PathParameter("workspace")
	r.namespaceName = req.PathParameter("namespace")
	r.workloadKind = req.PathParameter("kind")
	r.workloadName = req.PathParameter("workload")
	r.podName = req.PathParameter("pod")
	r.containerName = req.PathParameter("container")
	r.pvcName = req.PathParameter("pvc")
	r.storageClassName = req.PathParameter("storageclass")
	r.componentType = req.PathParameter("component")
	r.expression = req.QueryParameter("expr")
	r.metric = req.QueryParameter("metric")
	return r
}

func (h handler) makeQueryOptions(r reqParams, lvl monitoring.Level) (q queryOptions, err error) {
	if r.resourceFilter == "" {
		r.resourceFilter = DefaultFilter
	}

	q.metricFilter = r.metricFilter
	if r.metricFilter == "" {
		q.metricFilter = DefaultFilter
	}

	switch lvl {
	case monitoring.LevelCluster:
		q.option = monitoring.ClusterOption{}
		q.namedMetrics = model.ClusterMetrics
	case monitoring.LevelNode:
		q.identifier = model.IdentifierNode
		q.namedMetrics = model.NodeMetrics
		q.option = monitoring.NodeOption{
			ResourceFilter: r.resourceFilter,
			NodeName:       r.nodeName,
		}
	case monitoring.LevelWorkspace:
		q.identifier = model.IdentifierWorkspace
		q.namedMetrics = model.WorkspaceMetrics
		q.option = monitoring.WorkspaceOption{
			ResourceFilter: r.resourceFilter,
			WorkspaceName:  r.workspaceName,
		}
	case monitoring.LevelNamespace:
		q.identifier = model.IdentifierNamespace
		q.namedMetrics = model.NamespaceMetrics
		q.option = monitoring.NamespaceOption{
			ResourceFilter: r.resourceFilter,
			WorkspaceName:  r.workspaceName,
			NamespaceName:  r.namespaceName,
		}
	case monitoring.LevelWorkload:
		q.identifier = model.IdentifierWorkload
		q.namedMetrics = model.WorkloadMetrics
		q.option = monitoring.WorkloadOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
		}
	case monitoring.LevelPod:
		q.identifier = model.IdentifierPod
		q.namedMetrics = model.PodMetrics
		q.option = monitoring.PodOption{
			ResourceFilter: r.resourceFilter,
			NodeName:       r.nodeName,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
			WorkloadName:   r.workloadName,
			PodName:        r.podName,
		}
	case monitoring.LevelContainer:
		q.identifier = model.IdentifierContainer
		q.namedMetrics = model.ContainerMetrics
		q.option = monitoring.ContainerOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			PodName:        r.podName,
			ContainerName:  r.containerName,
		}
	case monitoring.LevelPVC:
		q.identifier = model.IdentifierPVC
		q.namedMetrics = model.PVCMetrics
		q.option = monitoring.PVCOption{
			ResourceFilter:            r.resourceFilter,
			NamespaceName:             r.namespaceName,
			StorageClassName:          r.storageClassName,
			PersistentVolumeClaimName: r.pvcName,
		}
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
	if r.namespaceName != "" {
		ns, err := h.k.CoreV1().Namespaces().Get(r.namespaceName, corev1.GetOptions{})
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
