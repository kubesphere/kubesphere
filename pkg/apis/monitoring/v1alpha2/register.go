/*

 Copyright 2019 The KubeSphere Authors.

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
package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/monitoring"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const GroupName = "monitoring.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	tags := []string{"Monitoring"}

	ws.Route(ws.GET("/clusters").To(monitoring.MonitorCluster).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("cluster_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes").To(monitoring.MonitorNode).
		Doc("monitor nodes level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("nodes_filter", "node re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}").To(monitoring.MonitorNode).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node", "specific node").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("node_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces").To(monitoring.MonitorNamespace).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_filter", "namespaces re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("namespace_memory_utilisation")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}").To(monitoring.MonitorNamespace).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(monitoring.MonitorPod).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(monitoring.MonitorPod).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods").To(monitoring.MonitorPod).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}").To(monitoring.MonitorPod).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod", "specific pod").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers").To(monitoring.MonitorContainer).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").Required(true)).
		Param(ws.PathParameter("pod", "specific pod").Required(true)).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").Required(false)).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers").To(monitoring.MonitorContainer).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").Required(false)).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(monitoring.MonitorContainer).
		Doc("monitor specific container level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").Required(true).DefaultValue("")).
		Param(ws.PathParameter("container", "specific container").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload_kind}").To(monitoring.MonitorWorkload).
		Doc("monitor specific workload level metrics").
		Param(ws.PathParameter("namespace", "namespace").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").Required(false)).
		Param(ws.PathParameter("workload_kind", "workload kind").Required(false).DefaultValue("daemonset")).
		Param(ws.QueryParameter("workload_name", "workload name").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "max metric items in a page").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/workloads").To(monitoring.MonitorWorkload).
		Doc("monitor all workload level metrics").
		Param(ws.PathParameter("namespace", "namespace").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").Required(false)).
		Param(ws.QueryParameter("workloads_filter", "pod re2 expression filter").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	// list all namespace in this workspace by selected metrics
	ws.Route(ws.GET("/workspaces/{workspace}").To(monitoring.MonitorOneWorkspace).
		Doc("monitor workspaces level metrics").
		Param(ws.PathParameter("workspace", "workspace name").Required(true)).
		Param(ws.QueryParameter("namespaces_filter", "namespaces filter").Required(false).DefaultValue("k.*")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("namespace_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/workspaces").To(monitoring.MonitorAllWorkspaces).
		Doc("monitor workspaces level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("workspace_memory_utilisation")).
		Param(ws.QueryParameter("workspaces_filter", "workspaces re2 expression filter").Required(false).DefaultValue(".*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").Required(false)).
		Param(ws.QueryParameter("page", "page number").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/components").To(monitoring.MonitorComponentStatus).
		Doc("monitor k8s components status").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	c.Add(ws)
	return nil
}
