package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"io"
	k8sinformers "k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	loggingclient "kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strconv"
	"strings"
)

const (
	levelCluster = iota
	levelWorkspace
	levelNamespace
	levelWorkload
	levelPod
	levelContainer

	// query type, default to `query`
	typeStat = "statistics"
	typeHist = "histogram"
	typeExport = "export"
)

type handler struct {
	informers k8sinformers.SharedInformerFactory
	c       loggingclient.Interface
}

func newHandler(k k8s.Client, l loggingclient.Interface) *handler {
	i := informers.NewInformerFactories(k.Kubernetes(), k.KubeSphere(), k.S2i(), k.Application()).KubernetesSharedInformerFactory()
	return &handler{i, l}
}

func (h handler) queryLog(req *restful.Request, resp *restful.Response) {
	typ := req.QueryParameter("type")

	noHit, sf := h.buildSearchFilter(levelCluster, req)
	if noHit {
		handleNoHit(typ, resp)
		return
	}

	switch typ {
	case typeStat:
		res, err := h.c.GetStatistics(sf)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	case typeHist:
		interval := req.QueryParameter("interval")
		res, err := h.c.CountLogsByInterval(sf, interval)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	case typeExport:
		resp.Header().Set(restful.HEADER_ContentType, "text/plain")
		resp.Header().Set("Content-Disposition", "attachment")
		b, err := h.c.ExportLogs(sf)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		_, err = io.Copy(resp, b)
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
		if order != loggingclient.Ascending {
			order = loggingclient.Descending
		}
		res, err := h.c.SearchLogs(sf, from, size, order)
		if err != nil {
			api.HandleInternalError(resp, err)
		}
		resp.WriteAsJson(res)
	}
}

func (h handler) buildSearchFilter(level int, req *restful.Request) (bool, loggingclient.SearchFilter) {
	var sf loggingclient.SearchFilter
	// if users try to query logs from namespaces they don't belong to,
	// return empty result directly without going through the logging store.
 	var noHit bool
	var namespaceFilter []string

	switch level {
	case levelCluster:
		noHit, namespaceFilter = logging.ListMatchedNamespaces(
			h.informers,
			stringutils.Split(req.QueryParameter("namespaces"), ","),
			stringutils.Split(strings.ToLower(req.QueryParameter("namespace_query")), ","),
			stringutils.Split(req.QueryParameter("workspaces"), ","),
			stringutils.Split(strings.ToLower(req.QueryParameter("workspace_query")), ","))
		sf.WorkloadFilter = stringutils.Split(req.QueryParameter("workloads"), ",")
		sf.WorkloadSearch = stringutils.Split(req.QueryParameter("workload_query"), ",")
		sf.PodFilter = stringutils.Split(req.QueryParameter("pods"), ",")
		sf.PodSearch = stringutils.Split(req.QueryParameter("pod_query"), ",")
		sf.ContainerFilter = stringutils.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = stringutils.Split(req.QueryParameter("container_query"), ",")
	case levelWorkspace:
		noHit, namespaceFilter = logging.ListMatchedNamespaces(
			h.informers,
			stringutils.Split(req.QueryParameter("namespaces"), ","),
			stringutils.Split(strings.ToLower(req.QueryParameter("namespace_query")), ","),
			stringutils.Split(req.PathParameter("workspace"), ","), nil)
		sf.WorkloadFilter = stringutils.Split(req.QueryParameter("workloads"), ",")
		sf.WorkloadSearch = stringutils.Split(req.QueryParameter("workload_query"), ",")
		sf.PodFilter = stringutils.Split(req.QueryParameter("pods"), ",")
		sf.PodSearch = stringutils.Split(req.QueryParameter("pod_query"), ",")
		sf.ContainerFilter = stringutils.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = stringutils.Split(req.QueryParameter("container_query"), ",")
	case levelNamespace:
		namespaceFilter = []string{req.PathParameter("namespace")}
		sf.WorkloadFilter = stringutils.Split(req.QueryParameter("workloads"), ",")
		sf.WorkloadSearch = stringutils.Split(req.QueryParameter("workload_query"), ",")
		sf.PodFilter = stringutils.Split(req.QueryParameter("pods"), ",")
		sf.PodSearch = stringutils.Split(req.QueryParameter("pod_query"), ",")
		sf.ContainerFilter = stringutils.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = stringutils.Split(req.QueryParameter("container_query"), ",")
	case levelWorkload:
		namespaceFilter = []string{req.PathParameter("namespace")}
		sf.WorkloadFilter = []string{req.PathParameter("workload")}
		sf.PodFilter = stringutils.Split(req.QueryParameter("pods"), ",")
		sf.PodSearch = stringutils.Split(req.QueryParameter("pod_query"), ",")
		sf.ContainerFilter = stringutils.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = stringutils.Split(req.QueryParameter("container_query"), ",")
	case levelPod:
		namespaceFilter = []string{req.PathParameter("namespace")}
		sf.PodFilter = []string{req.PathParameter("pod")}
		sf.ContainerFilter = stringutils.Split(req.QueryParameter("containers"), ",")
		sf.ContainerSearch = stringutils.Split(req.QueryParameter("container_query"), ",")
	case levelContainer:
		namespaceFilter = []string{req.PathParameter("namespace")}
		sf.PodFilter = []string{req.PathParameter("pod")}
		sf.ContainerFilter = []string{req.PathParameter("container")}
	}

	sf.NamespaceFilter = logging.WithCreationTimestamp(h.informers, namespaceFilter)
	sf.LogSearch = stringutils.Split(req.QueryParameter("log_query"), ",")
	sf.Starttime = req.QueryParameter("start_time")
	sf.Endtime = req.QueryParameter("end_time")

    return noHit, sf
}

func handleNoHit(typ string, resp *restful.Response) {
	switch typ {
	case typeStat:
		resp.WriteAsJson(new(loggingclient.Statistics))
	case typeHist:
		resp.WriteAsJson(new(loggingclient.Histogram))
	case typeExport:
		resp.Header().Set(restful.HEADER_ContentType, "text/plain")
		resp.Header().Set("Content-Disposition", "attachment")
		resp.Write(nil)
	default:
		resp.WriteAsJson(new(loggingclient.Logs))
	}
}
