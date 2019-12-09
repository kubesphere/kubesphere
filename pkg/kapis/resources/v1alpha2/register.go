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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/components"
	"kubesphere.io/kubesphere/pkg/apiserver/git"
	"kubesphere.io/kubesphere/pkg/apiserver/operations"
	"kubesphere.io/kubesphere/pkg/apiserver/quotas"
	"kubesphere.io/kubesphere/pkg/apiserver/registries"
	"kubesphere.io/kubesphere/pkg/apiserver/resources"
	"kubesphere.io/kubesphere/pkg/apiserver/revisions"
	"kubesphere.io/kubesphere/pkg/apiserver/routers"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/workloadstatuses"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	gitmodel "kubesphere.io/kubesphere/pkg/models/git"
	registriesmodel "kubesphere.io/kubesphere/pkg/models/registries"
	"kubesphere.io/kubesphere/pkg/models/status"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

const GroupName = "resources.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	webservice := runtime.NewWebService(GroupVersion)

	ok := "ok"

	webservice.Route(webservice.GET("/namespaces/{namespace}/{resources}").
		To(resources.ListNamespacedResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Namespace level resource query").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.PathParameter("resources", "namespace level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions,connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, models.PageableResponse{}))

	webservice.Route(webservice.POST("/namespaces/{namespace}/jobs/{job}").
		To(operations.RerunJob).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Rerun job whether the job is complete or not").
		Param(webservice.PathParameter("job", "job name")).
		Param(webservice.PathParameter("namespace", "the name of the namespace where the job runs in")).
		Param(webservice.QueryParameter("action", "action must be \"rerun\"")).
		Param(webservice.QueryParameter("resourceVersion", "version of job, rerun when the version matches")).
		Returns(http.StatusOK, ok, errors.Error{}))

	webservice.Route(webservice.GET("/{resources}").
		To(resources.ListResources).
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Doc("Cluster level resource query").
		Param(webservice.PathParameter("resources", "cluster level resource type, e.g. nodes,workspaces,storageclasses,clusterroles.")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.POST("/nodes/{node}/drainage").
		To(operations.DrainNode).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Doc("remove a node from service, safely evict all of your pods from a node and you can power down the node. More info: https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/").
		Param(webservice.PathParameter("node", "node name")).
		Returns(http.StatusOK, ok, errors.Error{}))

	webservice.Route(webservice.GET("/users/{user}/kubectl").
		To(resources.GetKubectl).
		Doc("get user's kubectl pod").
		Param(webservice.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourcesTag}).
		Returns(http.StatusOK, ok, models.PodInfo{}))

	webservice.Route(webservice.GET("/users/{user}/kubeconfig").
		Produces("text/plain", restful.MIME_JSON).
		To(resources.GetKubeconfig).
		Doc("get users' kubeconfig").
		Param(webservice.PathParameter("user", "username")).
		Returns(http.StatusOK, ok, "").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourcesTag}))

	webservice.Route(webservice.GET("/components").
		To(components.GetComponents).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("List the system components.").
		Returns(http.StatusOK, ok, []models.ComponentStatus{}))
	webservice.Route(webservice.GET("/components/{component}").
		To(components.GetComponentStatus).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("Describe the specified system component.").
		Param(webservice.PathParameter("component", "component name")).
		Returns(http.StatusOK, ok, models.ComponentStatus{}))
	webservice.Route(webservice.GET("/componenthealth").
		To(components.GetSystemHealthStatus).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("Get the health status of system components.").
		Returns(http.StatusOK, ok, models.HealthStatus{}))

	webservice.Route(webservice.GET("/quotas").
		To(quotas.GetClusterQuotas).
		Doc("get whole cluster's resource usage").
		Returns(http.StatusOK, ok, models.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/quotas").
		Doc("get specified namespace's resource quota and usage").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Returns(http.StatusOK, ok, models.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		To(quotas.GetNamespaceQuotas))

	webservice.Route(webservice.POST("registry/verify").
		To(registries.RegistryVerify).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VerificationTag}).
		Doc("verify if a user has access to the docker registry").
		Reads(registriesmodel.AuthInfo{}).
		Returns(http.StatusOK, ok, errors.Error{}))

	webservice.Route(webservice.GET("/registry/blob").
		To(registries.RegistryImageBlob).
		Param(webservice.QueryParameter("image", "query image, condition for filtering.").
			Required(true).
			DataFormat("image=%s")).
		Param(webservice.QueryParameter("namespace", "namespace which secret in.").
			Required(false).
			DataFormat("namespace=%s")).
		Param(webservice.QueryParameter("secret", "secret name").
			Required(false).
			DataFormat("secret=%s")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RegistryTag}).
		Doc("Retrieve the blob from the registry identified").
		Writes(registriesmodel.ImageDetails{}).
		Returns(http.StatusOK, ok, registriesmodel.ImageDetails{}),
	)

	webservice.Route(webservice.POST("git/verify").
		To(git.GitReadVerify).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VerificationTag}).
		Doc("Verify if the kubernetes secret has read access to the git repository").
		Reads(gitmodel.AuthInfo{}).
		Returns(http.StatusOK, ok, errors.Error{}),
	)

	webservice.Route(webservice.GET("/namespaces/{namespace}/daemonsets/{daemonset}/revisions/{revision}").
		To(revisions.GetDaemonSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified daemonset revision").
		Param(webservice.PathParameter("daemonset", "the name of the daemonset")).
		Param(webservice.PathParameter("namespace", "the namespace of the daemonset")).
		Param(webservice.PathParameter("revision", "the revision of the daemonset")).
		Returns(http.StatusOK, ok, appsv1.DaemonSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/deployments/{deployment}/revisions/{revision}").
		To(revisions.GetDeployRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified deployment revision").
		Param(webservice.PathParameter("deployment", "the name of deployment")).
		Param(webservice.PathParameter("namespace", "the namespace of the deployment")).
		Param(webservice.PathParameter("revision", "the revision of the deployment")).
		Returns(http.StatusOK, ok, appsv1.ReplicaSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/statefulsets/{statefulset}/revisions/{revision}").
		To(revisions.GetStatefulSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified statefulset revision").
		Param(webservice.PathParameter("statefulset", "the name of the statefulset")).
		Param(webservice.PathParameter("namespace", "the namespace of the statefulset")).
		Param(webservice.PathParameter("revision", "the revision of the statefulset")).
		Returns(http.StatusOK, ok, appsv1.StatefulSet{}))

	webservice.Route(webservice.GET("/routers").
		To(routers.GetAllRouters).
		Doc("List all routers of all projects").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Returns(http.StatusOK, ok, corev1.ServiceList{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/router").
		To(routers.GetRouter).
		Doc("List router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/router").
		To(routers.DeleteRouter).
		Doc("List router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.POST("/namespaces/{namespace}/router").
		To(routers.CreateRouter).
		Doc("Create a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/router").
		To(routers.UpdateRouter).
		Doc("Update a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.GET("/abnormalworkloads").
		Doc("get abnormal workloads' count of whole cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Returns(http.StatusOK, ok, status.WorkLoadStatus{}).
		To(workloadstatuses.GetClusterAbnormalWorkloads))
	webservice.Route(webservice.GET("/namespaces/{namespace}/abnormalworkloads").
		Doc("get abnormal workloads' count of specified namespace").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, status.WorkLoadStatus{}).
		To(workloadstatuses.GetNamespacedAbnormalWorkloads))

	c.Add(webservice)

	return nil
}
