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
	"kubesphere.io/kubesphere/pkg/apiserver/quotas"
	"kubesphere.io/kubesphere/pkg/apiserver/registries"
	"kubesphere.io/kubesphere/pkg/apiserver/resources"
	"kubesphere.io/kubesphere/pkg/apiserver/revisions"
	"kubesphere.io/kubesphere/pkg/apiserver/routers"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/workloadstatuses"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	gitmodel "kubesphere.io/kubesphere/pkg/models/git"
	"kubesphere.io/kubesphere/pkg/params"
)

const GroupName = "resources.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	webservice := runtime.NewWebService(GroupVersion)

	tags := []string{"Namespace resources"}

	webservice.Route(webservice.GET("/namespaces/{namespace}/{resources}").
		To(resources.NamespaceResourceHandler).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Namespace level resource query").
		Param(webservice.PathParameter("namespace", "which namespace")).
		Param(webservice.PathParameter("resources", "namespace level resource type")).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Writes(models.PageableResponse{}))

	tags = []string{"Cluster resources"}

	webservice.Route(webservice.GET("/{resources}").
		To(resources.ClusterResourceHandler).
		Writes(models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Cluster level resource query").
		Param(webservice.PathParameter("resources", "cluster level resource type"))).
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1"))

	webservice.Route(webservice.GET("/storageclasses/{storageclass}/persistentvolumeclaims").
		To(resources.GetPvcListBySc).
		Doc("get user's kubectl pod").
		Param(webservice.PathParameter("username", "username")).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	webservice.Route(webservice.GET("/namespaces/{namespace}/persistentvolumeclaims/{pvc}/pods").
		To(resources.GetPodListByPvc))

	tags = []string{"User resources"}

	webservice.Route(webservice.GET("/users/{username}/kubectl").
		To(resources.GetKubectl).
		Doc("get user's kubectl pod").
		Param(webservice.PathParameter("username", "username")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.PodInfo{}))

	webservice.Route(webservice.GET("/users/{username}/kubeconfig").
		Produces("text/plain").
		To(resources.GetKubeconfig).
		Doc("get users' kubeconfig").
		Param(webservice.PathParameter("username", "username")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	tags = []string{"Components"}

	webservice.Route(webservice.GET("/components").
		To(components.GetComponents).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Writes(map[string]models.Component{}))
	webservice.Route(webservice.GET("/components/{component}").
		To(components.GetComponentStatus).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Param(webservice.PathParameter("component", "component name")).
		Writes(models.Component{}))
	webservice.Route(webservice.GET("/health").
		To(components.GetSystemHealthStatus).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("").
		Writes(map[string]int{}))

	tags = []string{"Quotas"}

	webservice.Route(webservice.GET("/quotas").
		To(quotas.GetClusterQuotas).
		Doc("get whole cluster's resource usage").
		Writes(models.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	webservice.Route(webservice.GET("/namespaces/{namespace}/quotas").
		Doc("get specified namespace's resource quota and usage").
		Param(webservice.PathParameter("namespace", "namespace's name")).
		Writes(models.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(quotas.GetNamespaceQuotas))

	tags = []string{"Registries"}

	webservice.Route(webservice.POST("registries/verify").
		To(registries.RegistryVerify).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("docker registry verify").
		Writes(errors.Error{}))

	tags = []string{"Git"}
	webservice.Route(webservice.POST("/git/readverify").
		To(
			git.GitReadVerify).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("secret git read verify").
		Reads(gitmodel.AuthInfo{}).
		Writes(errors.Error{}),
	)
	tags = []string{"Revision"}

	webservice.Route(webservice.GET("/namespaces/{namespace}/daemonsets/{daemonset}/revisions/{revision}").
		To(revisions.GetDaemonSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Handle daemonset operation").
		Param(webservice.PathParameter("daemonset", "daemonset's name")).
		Param(webservice.PathParameter("namespace", "daemonset's namespace")).
		Param(webservice.PathParameter("revision", "daemonset's revision")).
		Writes(appsv1.DaemonSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/deployments/{deployment}/revisions/{revision}").
		To(revisions.GetDeployRevision).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Handle deployment operation").
		Param(webservice.PathParameter("deployment", "deployment's name")).
		Param(webservice.PathParameter("namespace", "deployment's namespace")).
		Param(webservice.PathParameter("revision", "deployment's revision")).
		Writes(appsv1.ReplicaSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/statefulsets/{statefulset}/revisions/{revision}").
		To(revisions.GetStatefulSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Doc("Handle statefulset operation").
		Param(webservice.PathParameter("statefulset", "statefulset's name")).
		Param(webservice.PathParameter("namespace", "statefulset's namespace")).
		Param(webservice.PathParameter("revision", "statefulset's revision")).
		Writes(appsv1.StatefulSet{}))

	tags = []string{"Router"}

	webservice.Route(webservice.GET("/routers").
		To(routers.GetAllRouters).
		Doc("Get all routers").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(corev1.Service{}))

	webservice.Route(webservice.GET("/users/{username}/routers").
		To(routers.GetAllRoutersOfUser).
		Doc("Get routers for user").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("username", "")).
		Writes(corev1.Service{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/router").
		To(routers.GetRouter).
		Doc("Get router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("namespace", "name of the project")))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/router").
		To(routers.DeleteRouter).
		Doc("Get router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("namespace", "name of the project")))

	webservice.Route(webservice.POST("/namespaces/{namespace}/router").
		To(routers.CreateRouter).
		Doc("Create a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("namespace", "name of the project")))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/router").
		To(routers.UpdateRouter).
		Doc("Update a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(webservice.PathParameter("namespace", "name of the project")))

	tags = []string{"WorkloadStatus"}

	webservice.Route(webservice.GET("/workloadstatuses").
		Doc("get abnormal workloads' count of whole cluster").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(workloadstatuses.GetClusterResourceStatus))
	webservice.Route(webservice.GET("/namespaces/{namespace}/workloadstatuses").
		Doc("get abnormal workloads' count of specified namespace").
		Param(webservice.PathParameter("namespace", "the name of namespace")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		To(workloadstatuses.GetNamespacesResourceStatus))

	c.Add(webservice)

	return nil
}
