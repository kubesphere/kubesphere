package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/servicemesh/metrics"
	"kubesphere.io/kubesphere/pkg/errors"
)

const GroupName = "servicemesh.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	tags := []string{"ServiceMesh"}

	webservice := runtime.NewWebService(GroupVersion)

	// Get service metrics
	// GET /namespaces/{namespace}/services/{service}/metrics
	webservice.Route(webservice.GET("/namespaces/{namespace}/services/{service}/metrics").
		To(metrics.GetServiceMetrics).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get app metrics from a specific namespace").
		Param(webservice.PathParameter("namespace", "name of the namespace")).
		Param(webservice.PathParameter("service", "name of the service")).
		Param(webservice.QueryParameter("filters[]", "type of metrics type, e.g. request_count, request_duration, request_error_count")).
		Param(webservice.QueryParameter("queryTime", "from which UNIX time to extract metrics")).
		Param(webservice.QueryParameter("duration", "metrics duration, in seconds")).
		Param(webservice.QueryParameter("step", "metrics step")).
		Param(webservice.QueryParameter("rateInterval", "metrics rate intervals, e.g. 20s")).
		Param(webservice.QueryParameter("quantiles[]", "metrics quantiles, 0.5, 0.9, 0.99")).
		Param(webservice.QueryParameter("byLabels[]", "by which labels to group node, e.g. source_workload, destination_service_name")).
		Param(webservice.QueryParameter("requestProtocol", "request protocol, http/tcp")).
		Param(webservice.QueryParameter("reporter", "destination")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get app metrics
	// Get /namespaces/{namespace}/apps/{app}/metrics
	webservice.Route(webservice.GET("/namespaces/{namespace}/apps/{app}/metrics").
		To(metrics.GetAppMetrics).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get app metrics from a specific namespace").
		Param(webservice.PathParameter("namespace", "name of the namespace")).
		Param(webservice.PathParameter("app", "name of the workload label app value")).
		Param(webservice.QueryParameter("filters[]", "type of metrics type, e.g. request_count, request_duration, request_error_count")).
		Param(webservice.QueryParameter("queryTime", "from which UNIX time to extract metrics")).
		Param(webservice.QueryParameter("duration", "metrics duration, in seconds")).
		Param(webservice.QueryParameter("step", "metrics step")).
		Param(webservice.QueryParameter("rateInterval", "metrics rate intervals, e.g. 20s")).
		Param(webservice.QueryParameter("quantiles[]", "metrics quantiles, 0.5, 0.9, 0.99")).
		Param(webservice.QueryParameter("byLabels[]", "by which labels to group node, e.g. source_workload, destination_service_name")).
		Param(webservice.QueryParameter("requestProtocol", "request protocol, http/tcp")).
		Param(webservice.QueryParameter("reporter", "destination")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get workload metrics
	// Get /namespaces/{namespace}/workloads/{workload}/metrics
	webservice.Route(webservice.GET("/namespaces/{namespace}/workloads/{workload}/metrics").
		To(metrics.GetWorkloadMetrics).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get workload metrics from a specific namespace").
		Param(webservice.PathParameter("namespace", "name of the namespace").Required(true)).
		Param(webservice.PathParameter("workload", "name of the workload").Required(true)).
		Param(webservice.QueryParameter("filters[]", "type of metrics type, e.g. request_count, request_duration, request_error_count")).
		Param(webservice.QueryParameter("queryTime", "from which UNIX time to extract metrics")).
		Param(webservice.QueryParameter("duration", "metrics duration, in seconds")).
		Param(webservice.QueryParameter("step", "metrics step")).
		Param(webservice.QueryParameter("rateInterval", "metrics rate intervals, e.g. 20s")).
		Param(webservice.QueryParameter("quantiles[]", "metrics quantiles, 0.5, 0.9, 0.99")).
		Param(webservice.QueryParameter("byLabels[]", "by which labels to group node, e.g. source_workload, destination_service_name")).
		Param(webservice.QueryParameter("requestProtocol", "request protocol, http/tcp")).
		Param(webservice.QueryParameter("reporter", "destination")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get namespace metrics
	// Get /namespaces/{namespace}/metrics
	webservice.Route(webservice.GET("/namespaces/{namespace}/metrics").
		To(metrics.GetNamespaceMetrics).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get workload metrics from a specific namespace").
		Param(webservice.PathParameter("namespace", "name of the namespace").Required(true)).
		Param(webservice.QueryParameter("filters[]", "type of metrics type, e.g. request_count, request_duration, request_error_count")).
		Param(webservice.QueryParameter("queryTime", "from which UNIX time to extract metrics")).
		Param(webservice.QueryParameter("duration", "metrics duration, in seconds")).
		Param(webservice.QueryParameter("step", "metrics step")).
		Param(webservice.QueryParameter("rateInterval", "metrics rate intervals, e.g. 20s")).
		Param(webservice.QueryParameter("quantiles[]", "metrics quantiles, 0.5, 0.9, 0.99")).
		Param(webservice.QueryParameter("byLabels[]", "by which labels to group node, e.g. source_workload, destination_service_name")).
		Param(webservice.QueryParameter("requestProtocol", "request protocol, http/tcp")).
		Param(webservice.QueryParameter("reporter", "destination")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get namespace graph
	// Get /namespaces/{namespace}/graph
	webservice.Route(webservice.GET("/namespaces/{namespace}/graph").
		To(metrics.GetNamespaceGraph).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get service graph for a specific namespace").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.QueryParameter("graphType", "type of the generated service graph, eg. ")).
		Param(webservice.QueryParameter("groupBy", "group nodes by kind")).
		Param(webservice.QueryParameter("queryTime", "from which time point, default now")).
		Param(webservice.QueryParameter("injectServiceNodes", "whether to inject service ndoes")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get namespaces graph, for multiple namespaces
	// Get /namespaces/graph
	webservice.Route(webservice.GET("/namespaces/{namespace}/graph").
		To(metrics.GetNamespacesGraph).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get service graph for a specific namespace").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.QueryParameter("graphType", "type of the generated service graph, eg. ")).
		Param(webservice.QueryParameter("groupBy", "group nodes by kind")).
		Param(webservice.QueryParameter("queryTime", "from which time point, default now")).
		Param(webservice.QueryParameter("injectServiceNodes", "whether to inject service ndoes")).
		Param(webservice.QueryParameter("namespaces", "names of namespaces")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get workloads health
	webservice.Route(webservice.GET("/namespaces/{namespace}/workloads/{workload}/health").
		To(metrics.GetWorkloadHealth).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get workload health").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.PathParameter("workload", "workload name").Required(true)).
		Param(webservice.QueryParameter("rateInterval", "the rate interval used for fetching error rate").DefaultValue("10m").Required(true)).
		Param(webservice.QueryParameter("queryTime", "the time to use for query")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get app health
	webservice.Route(webservice.GET("/namespaces/{namespace}/apps/{app}/health").
		To(metrics.GetAppHealth).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get workload health").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.PathParameter("app", "app name").Required(true)).
		Param(webservice.QueryParameter("rateInterval", "the rate interval used for fetching error rate").DefaultValue("10m").Required(true)).
		Param(webservice.QueryParameter("queryTime", "the time to use for query")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	// Get service health
	webservice.Route(webservice.GET("/namespaces/{namespace}/services/{service}/health").
		To(metrics.GetServiceHealth).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Get workload health").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.PathParameter("service", "service name").Required(true)).
		Param(webservice.QueryParameter("rateInterval", "the rate interval used for fetching error rate").DefaultValue("10m").Required(true)).
		Param(webservice.QueryParameter("queryTime", "the time to use for query")).
		Writes(errors.Error{})).Produces(restful.MIME_JSON)

	c.Add(webservice)

	return nil
}
