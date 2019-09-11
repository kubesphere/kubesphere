package metrics

import "net/url"

const (
	DefaultQueryStep    = "10m"
	DefaultQueryTimeout = "10s"
	RangeQueryType      = "query_range?"
	DefaultQueryType    = "query?"
)

type MonitoringRequestParams struct {
	Params           url.Values
	QueryType        string
	SortMetricName   string
	SortType         string
	PageNum          string
	LimitNum         string
	Tp               string
	MetricsFilter    string
	ResourcesFilter  string
	MetricsName      string
	WorkloadName     string
	NodeId           string
	WsName           string
	NsName           string
	PodName          string
	PVCName          string
	StorageClassName string
	ContainerName    string
	WorkloadKind     string
	ComponentName    string
}
