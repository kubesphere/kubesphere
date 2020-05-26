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
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	gitmodel "kubesphere.io/kubesphere/pkg/models/git"
	registriesmodel "kubesphere.io/kubesphere/pkg/models/registries"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

const (
	GroupName = "resources.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient kubernetes.Interface, factory informers.InformerFactory, masterURL string) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newResourceHandler(k8sClient, factory, masterURL)

	webservice.Route(webservice.GET("/namespaces/{namespace}/{resources}").
		To(handler.handleListNamespaceResources).
		Deprecate().
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
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}))

	webservice.Route(webservice.GET("/{resources}").
		To(handler.handleListNamespaceResources).
		Deprecate().
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Doc("Cluster level resources").
		Param(webservice.PathParameter("resources", "cluster level resource type, e.g. nodes,workspaces,storageclasses,clusterrole.")).
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

	webservice.Route(webservice.GET("/users/{user}/kubectl").
		To(handler.GetKubectlPod).
		Doc("get user's kubectl pod").
		Param(webservice.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, models.PodInfo{}))

	webservice.Route(webservice.GET("/users/{user}/kubeconfig").
		Produces("text/plain", restful.MIME_JSON).
		To(handler.GetKubeconfig).
		Doc("get users' kubeconfig").
		Param(webservice.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, "").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourcesTag}))

	webservice.Route(webservice.GET("/components").
		To(handler.handleGetComponents).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("List the system components.").
		Returns(http.StatusOK, api.StatusOK, []v1alpha2.ComponentStatus{}))

	webservice.Route(webservice.GET("/components/{component}").
		To(handler.handleGetComponentStatus).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("Describe the specified system component.").
		Param(webservice.PathParameter("component", "component name")).
		Returns(http.StatusOK, api.StatusOK, v1alpha2.ComponentStatus{}))
	webservice.Route(webservice.GET("/componenthealth").
		To(handler.handleGetSystemHealthStatus).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ComponentStatusTag}).
		Doc("Get the health status of system components.").
		Returns(http.StatusOK, api.StatusOK, v1alpha2.HealthStatus{}))

	webservice.Route(webservice.GET("/quotas").
		To(handler.handleGetClusterQuotas).
		Doc("get whole cluster's resource usage").
		Returns(http.StatusOK, api.StatusOK, api.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/quotas").
		Doc("get specified namespace's resource quota and usage").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Returns(http.StatusOK, api.StatusOK, api.ResourceQuota{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		To(handler.handleGetNamespaceQuotas))

	webservice.Route(webservice.POST("registry/verify").
		To(handler.handleVerifyRegistryCredential).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VerificationTag}).
		Doc("verify if a user has access to the docker registry").
		Reads(api.RegistryCredential{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))
	webservice.Route(webservice.GET("/registry/blob").
		To(handler.handleGetRegistryEntry).
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
		Returns(http.StatusOK, api.StatusOK, registriesmodel.ImageDetails{}),
	)
	webservice.Route(webservice.POST("git/verify").
		To(handler.handleVerifyGitCredential).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VerificationTag}).
		Doc("Verify if the kubernetes secret has read access to the git repository").
		Reads(gitmodel.AuthInfo{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}),
	)

	webservice.Route(webservice.GET("/namespaces/{namespace}/daemonsets/{daemonset}/revisions/{revision}").
		To(handler.handleGetDaemonSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified daemonset revision").
		Param(webservice.PathParameter("daemonset", "the name of the daemonset")).
		Param(webservice.PathParameter("namespace", "the namespace of the daemonset")).
		Param(webservice.PathParameter("revision", "the revision of the daemonset")).
		Returns(http.StatusOK, api.StatusOK, appsv1.DaemonSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/deployments/{deployment}/revisions/{revision}").
		To(handler.handleGetDeploymentRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified deployment revision").
		Param(webservice.PathParameter("deployment", "the name of deployment")).
		Param(webservice.PathParameter("namespace", "the namespace of the deployment")).
		Param(webservice.PathParameter("revision", "the revision of the deployment")).
		Returns(http.StatusOK, api.StatusOK, appsv1.ReplicaSet{}))
	webservice.Route(webservice.GET("/namespaces/{namespace}/statefulsets/{statefulset}/revisions/{revision}").
		To(handler.handleGetStatefulSetRevision).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Get the specified statefulset revision").
		Param(webservice.PathParameter("statefulset", "the name of the statefulset")).
		Param(webservice.PathParameter("namespace", "the namespace of the statefulset")).
		Param(webservice.PathParameter("revision", "the revision of the statefulset")).
		Returns(http.StatusOK, api.StatusOK, appsv1.StatefulSet{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/router").
		To(handler.handleGetRouter).
		Doc("List router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/router").
		To(handler.handleDeleteRouter).
		Doc("List router of a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.POST("/namespaces/{namespace}/router").
		To(handler.handleCreateRouter).
		Doc("Create a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/router").
		To(handler.handleUpdateRouter).
		Doc("Update a router for a specified project").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, corev1.Service{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.GET("/abnormalworkloads").
		Doc("get abnormal workloads' count of whole cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, api.Workloads{}).
		To(handler.handleGetNamespacedAbnormalWorkloads))
	webservice.Route(webservice.GET("/namespaces/{namespace}/abnormalworkloads").
		Doc("get abnormal workloads' count of specified namespace").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, api.Workloads{}).
		To(handler.handleGetNamespacedAbnormalWorkloads))

	c.Add(webservice)

	return nil
}
