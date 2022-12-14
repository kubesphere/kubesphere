package v1beta1

import (
	"net/http"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	tagClusteredResource  = "Clustered Resource"
	tagNamespacedResource = "Namespaced Resource"

	parameterGroup     = "group"
	parameterVersion   = "version"
	parameterResources = "resources"

	parameterNamespace = "namespace"

	ok = "OK"
)

var GroupVersion = schema.GroupVersion{Group: "resources", Version: "v1beta1"}

func AddToContainer(c *restful.Container, cache cache.Cache, cli client.Client, informerFactory informers.InformerFactory) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := New(v1beta1.New(cli, v1beta1.NewResourceCache(cache)), components.NewComponentsGetter(informerFactory.KubernetesSharedInformerFactory()))

	webservice.Route(webservice.GET("/{group}/{version}/{resources}").
		To(handler.listResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resource").
		Param(webservice.PathParameter(parameterGroup, "group name")).
		Param(webservice.PathParameter(parameterVersion, "version name")).
		Param(webservice.PathParameter(parameterResources, "resource name")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, nil))

	webservice.Route(webservice.GET("/{group}/{version}/{resources}/{name}").
		To(handler.getResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagClusteredResource}).
		Doc("Cluster level resource").
		Param(webservice.PathParameter(parameterGroup, "group name")).
		Param(webservice.PathParameter(parameterVersion, "version name")).
		Param(webservice.PathParameter(parameterResources, "resource name")).
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.GET("/namespace/{namespace}/{group}/{version}/{resources}").
		To(handler.listResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagNamespacedResource}).
		Doc("Namespace level resource").
		Param(webservice.PathParameter(parameterNamespace, "namespace name")).
		Param(webservice.PathParameter(parameterGroup, "group name")).
		Param(webservice.PathParameter(parameterVersion, "version name")).
		Param(webservice.PathParameter(parameterResources, "resource name")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, ok, nil))

	webservice.Route(webservice.GET("/namespace/{namespace}/{group}/{version}/{resources}/{name}").
		To(handler.getResources).
		Metadata(restfulspec.KeyOpenAPITags, []string{tagNamespacedResource}).
		Doc("Namespace level resource").
		Param(webservice.PathParameter(parameterNamespace, "namespace name")).
		Param(webservice.PathParameter(parameterGroup, "group name")).
		Param(webservice.PathParameter(parameterVersion, "version name")).
		Param(webservice.PathParameter(parameterResources, "resource name")).
		Returns(http.StatusOK, api.StatusOK, nil))

	c.Add(webservice)
	return nil
}
