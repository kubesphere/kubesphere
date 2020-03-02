package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/models/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	util "kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strconv"
	"strings"
	"time"
)

const (
	LevelCluster = iota
	LevelContainer

	// query type, default to `query`
	TypeStat   = "statistics"
	TypeHist   = "histogram"
	TypeExport = "export"

	Ascending  = "asc"
	Descending = "desc"
)

type handler struct {
	k  k8s.Client
	lo logging.LoggingOperator
}

func newHandler(k k8s.Client, l loggingclient.Interface) *handler {
	return &handler{k, logging.NewLoggingOperator(l)}
}

func (h handler) handleClusterQuery(req *restful.Request, resp *restful.Response) {
	h.get(req, LevelCluster, resp)
}

func (h handler) handleContainerQuery(req *restful.Request, resp *restful.Response) {
	h.get(req, LevelContainer, resp)
}

func (h handler) get(req *restful.Request, lvl int, resp *restful.Response) {
	typ := req.QueryParameter("type")

	noHit, sf, err := h.newSearchFilter(req, lvl)
	if err != nil {
		api.HandleBadRequest(resp, err)
	}
	if noHit {
		handleNoHit(typ, resp)
		return
	}

	switch typ {
	case TypeStat:
		res, err := h.lo.GetCurrentStats(sf)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	case TypeHist:
		interval := req.QueryParameter("interval")
		res, err := h.lo.CountLogsByInterval(sf, interval)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	case TypeExport:
		resp.Header().Set(restful.HEADER_ContentType, "text/plain")
		resp.Header().Set("Content-Disposition", "attachment")
		err := h.lo.ExportLogs(sf, resp.ResponseWriter)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
	default:
		from, _ := strconv.ParseInt(req.QueryParameter("from"), 10, 64)
		size, err := strconv.ParseInt(req.QueryParameter("size"), 10, 64)
		if err != nil {
			size = 10
		}
		order := req.QueryParameter("sort")
		if order != Ascending {
			order = Descending
		}
		res, err := h.lo.SearchLogs(sf, from, size, order)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	}
}

func (h handler) newSearchFilter(req *restful.Request, level int) (bool, loggingclient.SearchFilter, error) {
	var sf loggingclient.SearchFilter

	switch level {
	case LevelCluster:
		sf.NamespaceFilter = h.intersect(
			util.Split(req.QueryParameter("namespaces"), ","),
			util.Split(strings.ToLower(req.QueryParameter("namespace_query")), ","),
			util.Split(req.QueryParameter("workspaces"), ","),
			util.Split(strings.ToLower(req.QueryParameter("workspace_query")), ","))
		sf.WorkloadFilter = util.Split(req.QueryParameter("workloads"), ",")
		sf.WorkloadSearch = util.Split(req.QueryParameter("workload_query"), ",")
		sf.PodFilter = util.Split(req.QueryParameter("pods"), ",")
		sf.PodSearch = util.Split(req.QueryParameter("pod_query"), ",")
		sf.ContainerFilter = util.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = util.Split(req.QueryParameter("container_query"), ",")
	case LevelContainer:
		sf.NamespaceFilter = h.withCreationTime(req.PathParameter("namespace"))
		sf.PodFilter = []string{req.PathParameter("pod")}
		sf.ContainerFilter = []string{req.PathParameter("container")}
	}

	sf.LogSearch = util.Split(req.QueryParameter("log_query"), ",")

	var err error
	now := time.Now()
	// If time is not given, set it to now.
	if req.QueryParameter("start_time") == "" {
		sf.Starttime = now
	} else {
		sf.Starttime, err = time.Parse(time.RFC3339, req.QueryParameter("start_time"))
		if err != nil {
			return false, sf, err
		}
	}
	if req.QueryParameter("end_time") == "" {
		sf.Endtime = now
	} else {
		sf.Endtime, err = time.Parse(time.RFC3339, req.QueryParameter("end_time"))
		if err != nil {
			return false, sf, err
		}
	}

	return len(sf.NamespaceFilter) == 0, sf, nil
}

func handleNoHit(typ string, resp *restful.Response) {
	switch typ {
	case TypeStat:
		resp.WriteAsJson(new(loggingclient.Statistics))
	case TypeHist:
		resp.WriteAsJson(new(loggingclient.Histogram))
	case TypeExport:
		resp.Header().Set(restful.HEADER_ContentType, "text/plain")
		resp.Header().Set("Content-Disposition", "attachment")
		resp.Write(nil)
	default:
		resp.WriteAsJson(new(loggingclient.Logs))
	}
}
