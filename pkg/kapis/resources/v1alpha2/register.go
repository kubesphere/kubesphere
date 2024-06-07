/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/git"
	gitmodel "kubesphere.io/kubesphere/pkg/models/git"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/quotas"
	"kubesphere.io/kubesphere/pkg/models/registries"
	registriesmodel "kubesphere.io/kubesphere/pkg/models/registries"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/models/revisions"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "resources.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func NewHandler(cacheClient runtimeclient.Client, k8sVersion *semver.Version, masterURL string, options *terminal.Options) rest.Handler {
	return &handler{
		resourceGetter:      resourcev1alpha3.NewResourceGetter(cacheClient, k8sVersion),
		componentsGetter:    components.NewComponentsGetter(cacheClient),
		resourceQuotaGetter: quotas.NewResourceQuotaGetter(cacheClient, k8sVersion),
		revisionGetter:      revisions.NewRevisionGetter(cacheClient),
		gitVerifier:         git.NewGitVerifier(cacheClient),
		registryGetter:      registries.NewRegistryGetter(cacheClient),
		kubeconfigOperator:  kubeconfig.NewReadOnlyOperator(cacheClient, masterURL),
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/users/{user}/kubeconfig").
		Produces("text/plain", restful.MIME_JSON).
		To(h.GetKubeconfig).
		Deprecate().
		Doc("Get users' kubeconfig").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, ""))

	ws.Route(ws.GET("/components").
		To(h.GetComponents).
		Deprecate().
		Doc("List the system components").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Operation("get-components-v1alpha2").
		Returns(http.StatusOK, api.StatusOK, []v1alpha2.ComponentStatus{}))

	ws.Route(ws.GET("/components/{component}").
		To(h.GetComponentStatus).
		Deprecate().
		Doc("Describe the specified system component").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Operation("get-components-status-v1alpha2").
		Param(ws.PathParameter("component", "component name")).
		Returns(http.StatusOK, api.StatusOK, v1alpha2.ComponentStatus{}))
	ws.Route(ws.GET("/componenthealth").
		To(h.GetSystemHealthStatus).
		Deprecate().
		Doc("Get the health status of system components").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Operation("get-system-health-status-v1alpha2").
		Returns(http.StatusOK, api.StatusOK, v1alpha2.HealthStatus{}))

	ws.Route(ws.GET("/quotas").
		To(h.GetClusterQuotas).
		Deprecate().
		Doc("Get whole cluster's resource usage").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagClusterResources}).
		Returns(http.StatusOK, api.StatusOK, api.ResourceQuota{}))

	ws.Route(ws.GET("/namespaces/{namespace}/quotas").
		To(h.GetNamespaceQuotas).
		Deprecate().
		Doc("get specified namespace's resource quota and usage").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Returns(http.StatusOK, api.StatusOK, api.ResourceQuota{}))

	ws.Route(ws.POST("registry/verify").
		To(h.VerifyRegistryCredential).
		Deprecate().
		Doc("Verify registry credential").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAdvancedOperations}).
		Reads(api.RegistryCredential{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))
	ws.Route(ws.GET("/registry/blob").
		To(h.GetRegistryEntry).
		Deprecate().
		Doc("Retrieve the blob from the registry").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAdvancedOperations}).
		Param(ws.QueryParameter("image", "query image, condition for filtering.").
			Required(true).
			DataFormat("image=%s")).
		Param(ws.QueryParameter("namespace", "namespace which secret in.").
			Required(false).
			DataFormat("namespace=%s")).
		Param(ws.QueryParameter("secret", "secret name").
			Required(false).
			DataFormat("secret=%s")).
		Param(ws.QueryParameter("insecure", "whether verify cert if using https repo").
			Required(false).
			DataFormat("insecure=%s")).
		Writes(registriesmodel.ImageDetails{}).
		Returns(http.StatusOK, api.StatusOK, registriesmodel.ImageDetails{}))
	ws.Route(ws.POST("git/verify").
		To(h.VerifyGitCredential).
		Deprecate().
		Doc("Verify the git credential").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAdvancedOperations}).
		Reads(gitmodel.AuthInfo{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}),
	)

	ws.Route(ws.GET("/namespaces/{namespace}/daemonsets/{daemonset}/revisions/{revision}").
		To(h.GetDaemonSetRevision).
		Deprecate().
		Doc("Get the specified daemonSet revision").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("daemonset", "the name of the daemonset")).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("revision", "the revision of the daemonset")).
		Returns(http.StatusOK, api.StatusOK, appsv1.DaemonSet{}))
	ws.Route(ws.GET("/namespaces/{namespace}/deployments/{deployment}/revisions/{revision}").
		To(h.GetDeploymentRevision).
		Deprecate().
		Doc("Get the specified deployment revision").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("deployment", "the name of deployment")).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("revision", "the revision of the deployment")).
		Returns(http.StatusOK, api.StatusOK, appsv1.ReplicaSet{}))
	ws.Route(ws.GET("/namespaces/{namespace}/statefulsets/{statefulset}/revisions/{revision}").
		To(h.GetStatefulSetRevision).
		Deprecate().
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Doc("Get the specified statefulSet revision").
		Param(ws.PathParameter("statefulset", "the name of the statefulset")).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("revision", "the revision of the statefulset")).
		Returns(http.StatusOK, api.StatusOK, appsv1.StatefulSet{}))

	ws.Route(ws.GET("/abnormalworkloads").
		To(h.GetNamespacedAbnormalWorkloads).
		Deprecate().
		Doc("Get abnormal workloads").
		Operation("get-cluster-abnormal-workloads").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagClusterResources}).
		Returns(http.StatusOK, api.StatusOK, api.Workloads{}))
	ws.Route(ws.GET("/namespaces/{namespace}/abnormalworkloads").
		To(h.GetNamespacedAbnormalWorkloads).
		Deprecate().
		Doc("Get abnormal workloads").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Operation("get-namespaced-abnormal-workloads").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Returns(http.StatusOK, api.StatusOK, api.Workloads{}))

	c.Add(ws)
	return nil
}
