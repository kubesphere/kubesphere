package v1alpha1

import (
	"time"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	DefaultStep   = 10 * time.Minute
	DefaultFilter = ".*"
	DefaultOrder  = model.OrderDescending
	DefaultPage   = 1
	DefaultLimit  = 5

	ErrNoHit             = "'end' or 'time' must be after the namespace creation time."
	ErrParamConflict     = "'time' and the combination of 'start' and 'end' are mutually exclusive."
	ErrInvalidStartEnd   = "'start' must be before 'end'."
	ErrInvalidPage       = "Invalid parameter 'page'."
	ErrInvalidLimit      = "Invalid parameter 'limit'."
	ErrParameterNotfound = "Parmameter [%s] not found"
	ErrResourceNotfound  = "resource not found"
	ErrScopeNotAllowed   = "scope [%s] not allowed"
)

type Query struct {
	Level            monitoring.Level
	Operation        string
	LabelSelector    string
	Time             string
	Start            string
	End              string
	Step             string
	Target           string
	Order            string
	Page             string
	Limit            string
	MetricFilter     string
	ResourceFilter   string
	NodeName         string
	WorkspaceName    string
	NamespaceName    string
	WorkloadKind     string
	WorkloadName     string
	PodName          string
	Applications     string
	Services         string
	StorageClassName string
	PVCFilter        string
}

func ParseQueryParameter(req *restful.Request) *Query {
	var q Query

	q.LabelSelector = req.QueryParameter(query.ParameterLabelSelector)

	q.Level = monitoring.Level(monitoring.MeteringLevelMap[req.QueryParameter("level")])
	q.Operation = req.QueryParameter("operation")
	q.Time = req.QueryParameter("time")
	q.Start = req.QueryParameter("start")
	q.End = req.QueryParameter("end")
	q.Step = req.QueryParameter("step")
	q.Target = req.QueryParameter("sort_metric")
	q.Order = req.QueryParameter("sort_type")
	q.Page = req.QueryParameter("page")
	q.Limit = req.QueryParameter("limit")
	q.MetricFilter = req.QueryParameter("metrics_filter")
	q.ResourceFilter = req.QueryParameter("resources_filter")
	q.WorkspaceName = req.QueryParameter("workspace")

	q.NamespaceName = req.QueryParameter("namespace")
	if q.NamespaceName == "" {
		q.NamespaceName = req.PathParameter("namespace")
	}

	q.NodeName = req.QueryParameter("node")
	q.WorkloadKind = req.QueryParameter("kind")
	q.WorkloadName = req.QueryParameter("workload")
	q.PodName = req.QueryParameter("pod")
	q.Applications = req.QueryParameter("applications")
	q.Services = req.QueryParameter("services")
	q.StorageClassName = req.QueryParameter("storageclass")
	q.PVCFilter = req.QueryParameter("pvc_filter")

	return &q
}
