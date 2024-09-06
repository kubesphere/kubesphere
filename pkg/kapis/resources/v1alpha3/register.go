/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha3

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/components"
	v2 "kubesphere.io/kubesphere/pkg/models/registries/v2"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/overview"
)

const (
	GroupName = "resources.kubesphere.io"
	Version   = "v1alpha3"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func NewHandler(cacheReader runtimeclient.Reader, counter overview.Counter, k8sVersion *semver.Version) rest.Handler {
	return &handler{
		resourceGetterV1alpha3: resourcev1alpha3.NewResourceGetter(cacheReader, k8sVersion),
		componentsGetter:       components.NewComponentsGetter(cacheReader),
		registryHelper:         v2.NewRegistryHelper(),
		counter:                counter,
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/{resources}").
		To(h.ListResources).
		Deprecate().
		Doc("Cluster level resources").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagClusterResources}).
		Operation("list-cluster-resources").
		Param(ws.PathParameter("resources", "cluster level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}))

	ws.Route(ws.GET("/{resources}/{name}").
		To(h.GetResources).
		Deprecate().
		Doc("Get cluster scope resource").
		Operation("get-cluster-resource").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagClusterResources}).
		Param(ws.PathParameter("resources", "cluster level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(ws.PathParameter("name", "the name of the clustered resources")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/namespaces/{namespace}/{resources}").
		To(h.ListResources).
		Deprecate().
		Doc("List resources at namespace scope").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Operation("list-namespaced-resources").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("resources", "namespace level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(ws.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Param(ws.QueryParameter(query.ParameterFieldSelector, "field selector used for filtering, you can use the = , == and != operators with field selectors( = and == mean the same thing), e.g. fieldSelector=type=kubernetes.io/dockerconfigjson, multiple separated by comma").Required(false)).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}))

	ws.Route(ws.GET("/namespaces/{namespace}/{resources}/{name}").
		To(h.GetResources).
		Deprecate().
		Doc("Get namespace scope resource").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Operation("get-namespaced-resource").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("resources", "namespace level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(ws.PathParameter("name", "the name of resource")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}))

	ws.Route(ws.GET("/components").
		To(h.GetComponents).
		Deprecate().
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Doc("List the system components").
		Operation("get-components-v1alpha3").
		Returns(http.StatusOK, api.StatusOK, []v1alpha2.ComponentStatus{}))
	ws.Route(ws.GET("/components/{component}").
		To(h.GetComponentStatus).
		Deprecate().
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Doc("Describe the specified system component").
		Operation("get-components-status-v1alpha3").
		Param(ws.PathParameter("component", "component name")).
		Returns(http.StatusOK, api.StatusOK, v1alpha2.ComponentStatus{}))
	ws.Route(ws.GET("/componenthealth").
		To(h.GetSystemHealthStatus).
		Deprecate().
		Doc("Get the health status of system components").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagComponentStatus}).
		Operation("get-system-health-status-v1alpha3").
		Returns(http.StatusOK, api.StatusOK, v1alpha2.HealthStatus{}))

	ws.Route(ws.POST("/namespaces/{namespace}/registrysecrets/{secret}/verify").
		To(h.VerifyImageRepositorySecret).
		Deprecate().
		Doc("Verify registry credential").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("secret", "Name of the secret.")).
		Reads(v1.Secret{}).
		Returns(http.StatusOK, api.StatusOK, v1.Secret{}))

	ws.Route(ws.GET("/namespaces/{namespace}/imageconfig").
		To(h.GetImageConfig).
		Deprecate().
		Doc("Get image config").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.QueryParameter("secret", "Secret name of the image repository credential, left empty means anonymous fetch.").Required(false)).
		Param(ws.QueryParameter("image", "Image name to query, e.g. kubesphere/ks-apiserver:v3.1.1").Required(true)).
		Returns(http.StatusOK, api.StatusOK, v2.ImageConfig{}))

	ws.Route(ws.GET("/namespaces/{namespace}/repositorytags").
		To(h.GetRepositoryTags).
		Deprecate().
		Doc("List image tags").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Notes("List repository tags, this is an experimental API, use it by your own caution.").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.QueryParameter("repository", "Repository to query, e.g. calico/cni.").Required(true)).
		Param(ws.QueryParameter("secret", "Secret name of the image repository credential, left empty means anonymous fetch.").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Returns(http.StatusOK, api.StatusOK, v2.RepositoryTags{}))

	ws.Route(ws.GET("/metrics").
		To(h.GetClusterOverview).
		Deprecate().
		Doc("Cluster summary").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagClusterResources}).
		Returns(http.StatusOK, api.StatusOK, overview.MetricResults{}))

	ws.Route(ws.GET("/namespaces/{namespace}/metrics").
		To(h.GetNamespaceOverview).
		Deprecate().
		Doc("Namespace summary").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagNamespacedResources}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Returns(http.StatusOK, api.StatusOK, overview.MetricResults{}))

	c.Add(ws)
	return nil
}
