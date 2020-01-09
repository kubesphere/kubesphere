package v1alpha2

import (
	"fmt"
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
)

type params struct {
	time       time.Time
	start, end time.Time
	step       time.Duration

	target     string
	identifier string
	order      string
	page       int
	limit      int

	option monitoring.QueryOption
}

func (p params) isRangeQuery() bool {
	return !p.time.IsZero()
}

func (p params) shouldSort() bool {
	return p.target != ""
}

func (h handler) parseRequestParams(req *restful.Request, lvl monitoring.MonitoringLevel) (params, error) {
	timestamp := req.QueryParameter("time")
	start := req.QueryParameter("start")
	end := req.QueryParameter("end")
	step := req.QueryParameter("step")
	target := req.QueryParameter("sort_metric")
	order := req.QueryParameter("sort_type")
	page := req.QueryParameter("page")
	limit := req.QueryParameter("limit")
	metricFilter := req.QueryParameter("metrics_filter")
	resourceFilter := req.QueryParameter("resources_filter")
	nodeName := req.PathParameter("node")
	workspaceName := req.PathParameter("workspace")
	namespaceName := req.PathParameter("namespace")
	workloadKind := req.PathParameter("kind")
	workloadName := req.PathParameter("workload")
	podName := req.PathParameter("pod")
	containerName := req.PathParameter("container")
	pvcName := req.PathParameter("pvc")
	storageClassName := req.PathParameter("storageclass")
	componentType := req.PathParameter("component")

	var p params
	var err error
	if start != "" && end != "" {
		p.start, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return p, err
		}
		p.end, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return p, err
		}
		if step == "" {
			p.step = DefaultStep
		} else {
			p.step, err = time.ParseDuration(step)
			if err != nil {
				return p, err
			}
		}
	} else if start == "" && end == "" {
		if timestamp == "" {
			p.time = time.Now()
		} else {
			p.time, err = time.Parse(time.RFC3339, req.QueryParameter("time"))
			if err != nil {
				return p, err
			}
		}
	} else {
		return p, errors.Errorf("'time' and the combination of 'start' and 'end' are mutually exclusive.")
	}

	// hide metrics from a deleted namespace having the same name
	namespace := req.QueryParameter("namespace")
	if req.QueryParameter("namespace") != "" {
		ns, err := h.k.Kubernetes().CoreV1().Namespaces().Get(namespace, corev1.GetOptions{})
		if err != nil {
			return p, err
		}

		cts := ns.CreationTimestamp.Time
		if p.start.Before(cts) {
			p.start = cts
		}
		if p.end.Before(cts) {
			return p, errors.Errorf("End timestamp must not be before namespace creation time.")
		}
	}

	if resourceFilter == "" {
		resourceFilter = DefaultFilter
	}

	if metricFilter == "" {
		metricFilter = DefaultFilter
	}
	if componentType != "" {
		metricFilter = fmt.Sprintf("/^(?=.*%s)(?=.*%s)/s", componentType, metricFilter)
	}

	// should sort
	if target != "" {
		p.page = DefaultPage
		p.limit = DefaultLimit
		if order != model.OrderAscending {
			p.order = DefaultOrder
		}
		if page != "" {
			p.page, err = strconv.Atoi(req.QueryParameter("page"))
			if err != nil || p.page <= 0 {
				return p, errors.Errorf("Invalid parameter 'page'.")
			}
		}
		if limit != "" {
			p.limit, err = strconv.Atoi(req.QueryParameter("limit"))
			if err != nil || p.limit <= 0 {
				return p, errors.Errorf("Invalid parameter 'limit'.")
			}
		}
	}

	switch lvl {
	case monitoring.LevelCluster:
		p.option = monitoring.ClusterOption{MetricFilter: metricFilter}
	case monitoring.LevelNode:
		p.identifier = model.IdentifierNode
		p.option = monitoring.NodeOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			NodeName:       nodeName,
		}
	case monitoring.LevelWorkspace:
		p.identifier = model.IdentifierWorkspace
		p.option = monitoring.WorkspaceOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			WorkspaceName:  workspaceName,
		}
	case monitoring.LevelNamespace:
		p.identifier = model.IdentifierNamespace
		p.option = monitoring.NamespaceOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			WorkspaceName:  workspaceName,
			NamespaceName:  namespaceName,
		}
	case monitoring.LevelWorkload:
		p.identifier = model.IdentifierWorkload
		p.option = monitoring.WorkloadOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			NamespaceName:  namespaceName,
			WorkloadKind:   workloadKind,
			WorkloadName:   workloadName,
		}
	case monitoring.LevelPod:
		p.identifier = model.IdentifierPod
		p.option = monitoring.PodOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			NodeName:       nodeName,
			NamespaceName:  namespaceName,
			WorkloadKind:   workloadKind,
			WorkloadName:   workloadName,
			PodName:        podName,
		}
	case monitoring.LevelContainer:
		p.identifier = model.IdentifierContainer
		p.option = monitoring.ContainerOption{
			MetricFilter:   metricFilter,
			ResourceFilter: resourceFilter,
			NamespaceName:  namespaceName,
			PodName:        podName,
			ContainerName:  containerName,
		}
	case monitoring.LevelPVC:
		p.identifier = model.IdentifierPVC
		p.option = monitoring.PVCOption{
			MetricFilter:              metricFilter,
			ResourceFilter:            resourceFilter,
			NamespaceName:             namespaceName,
			StorageClassName:          storageClassName,
			PersistentVolumeClaimName: pvcName,
		}
	case monitoring.LevelComponent:
		p.option = monitoring.ComponentOption{
			MetricFilter: metricFilter,
		}
	}

	return p, nil
}
