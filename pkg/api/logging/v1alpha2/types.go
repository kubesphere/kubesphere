package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"strconv"
	"time"
)

const (
	OperationStatistics = "statistics"
	OperationHistogram  = "histogram"
	OperationQuery      = "query"
	OperationExport     = "export"

	DefaultInterval = "15m"
	DefaultSize     = 10
	OrderAscending  = "asc"
	OrderDescending = "desc"
)

type APIResponse struct {
	Logs       *logging.Logs       `json:"query,omitempty" description:"query results"`
	Statistics *logging.Statistics `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *logging.Histogram  `json:"histogram,omitempty" description:"histogram results"`
}

type Query struct {
	Operation       string
	WorkspaceFilter string
	WorkspaceSearch string
	NamespaceFilter string
	NamespaceSearch string
	WorkloadFilter  string
	WorkloadSearch  string
	PodFilter       string
	PodSearch       string
	ContainerFilter string
	ContainerSearch string
	LogSearch       string
	StartTime       time.Time
	EndTime         time.Time
	Interval        string
	Sort            string
	From            int64
	Size            int64
}

func ParseQueryParameter(req *restful.Request) (*Query, error) {
	var q Query
	q.Operation = req.QueryParameter("operation")
	q.WorkspaceFilter = req.QueryParameter("workspaces")
	q.WorkspaceSearch = req.QueryParameter("workspace_query")
	q.NamespaceFilter = req.QueryParameter("namespaces")
	q.NamespaceSearch = req.QueryParameter("namespace_query")
	q.WorkloadFilter = req.QueryParameter("workloads")
	q.WorkloadSearch = req.QueryParameter("workload_query")
	q.PodFilter = req.QueryParameter("pods")
	q.PodSearch = req.QueryParameter("pod_query")
	q.ContainerFilter = req.QueryParameter("containers")
	q.ContainerSearch = req.QueryParameter("container_query")
	q.LogSearch = req.QueryParameter("log_query")

	if q.Operation == "" {
		q.Operation = OperationQuery
	}

	if tstr := req.QueryParameter("start_time"); tstr != "" {
		sec, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return nil, err
		}
		q.StartTime = time.Unix(sec, 0)
	}
	if tstr := req.QueryParameter("end_time"); tstr != "" {
		sec, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return nil, err
		}
		q.EndTime = time.Unix(sec, 0)
	}

	switch q.Operation {
	case OperationHistogram:
		q.Interval = req.QueryParameter("interval")
		if q.Interval == "" {
			q.Interval = DefaultInterval
		}
	case OperationQuery:
		q.From, _ = strconv.ParseInt(req.QueryParameter("from"), 10, 64)
		size, err := strconv.ParseInt(req.QueryParameter("size"), 10, 64)
		if err != nil {
			size = DefaultSize
		}
		q.Size = size
		q.Sort = req.QueryParameter("sort")
		if q.Sort != OrderAscending {
			q.Sort = OrderDescending
		}
	}

	return &q, nil
}
