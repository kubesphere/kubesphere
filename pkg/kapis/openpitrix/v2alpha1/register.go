/*
Copyright 2020 The KubeSphere Authors.
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

package v2alpha1

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/params"
	openpitrixoptions "kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"net/http"
)

const (
	GroupName = "openpitrix.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v2alpha1"}

func AddToContainer(c *restful.Container, ksInfomrers informers.InformerFactory, ksClient versioned.Interface, options *openpitrixoptions.Options) error {
	webservice := runtime.NewWebService(GroupVersion)

	handler := newOpenpitrixHandler(ksInfomrers, ksClient, options)

	webservice.Route(webservice.GET("/workspaces/{workspace}/repos").
		To(handler.ListRepos).
		Doc("List repositories in the specified workspace").
		Param(webservice.PathParameter("workspace", "the name of the workspace.").Required(true)).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}))

	webservice.Route(webservice.GET("/workspaces/{workspace}/repos/{repo}").
		To(handler.DescribeRepo).
		Doc("Describe the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmRepo{}).
		Param(webservice.PathParameter("repo", "repo id")))

	webservice.Route(webservice.GET("/applications").
		Deprecate().
		To(handler.ListApplications).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Doc("List all applications").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications").
		To(handler.ListApplications).
		Doc("List all applications within the specified namespace").
		Param(webservice.PathParameter("namespace", "the name of the namespace.").Required(true)).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/applications").
		To(handler.ListApplications).
		Param(webservice.PathParameter("workspace", "the name of the workspace.").Required(true)).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Doc("List all applications within the specified workspace").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/applications/{application}").
		To(handler.DescribeApplication).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmRelease{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.GET("/workspaces/{workspace}/applications/{application} ").
		To(handler.DescribeApplication).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmRelease{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.GET("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications/{application}").
		To(handler.DescribeApplication).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmRelease{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.GET("/apps").
		To(handler.ListApps).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Doc("List all apps").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/apps").
		To(handler.ListApps).
		Param(webservice.PathParameter("workspace", "the name of the workspace.").Required(true)).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Doc("List all apps within the specified workspace").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/apps/{app}").
		To(handler.DescribeApp).
		Doc("Describe the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmApplication{}).
		Param(webservice.PathParameter("app", "app template id")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}").
		To(handler.DescribeApp).
		Doc("Describe the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmApplication{}).
		Param(webservice.PathParameter("app", "app template id")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}/versions").
		To(handler.ListAppVersion).
		Param(webservice.PathParameter("workspace", "the name of the workspace.").Required(true)).
		Param(webservice.PathParameter("app", "the id of the app.").Required(true)).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Doc("List all apps within the specified workspace").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/apps/{app}/versions").
		To(handler.ListAppVersion).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmRepo{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Param(webservice.PathParameter("app", "the id of the app.").Required(true)).
		Doc("List all apps").
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}/versions/{version}").
		To(handler.DescribeAppVersion).
		Doc("Describe the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmApplication{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")).
		Param(webservice.PathParameter("workspaces", "the name of the workspace")))

	webservice.Route(webservice.GET("/apps/{app}/versions/{version}").
		To(handler.DescribeAppVersion).
		Doc("Describe the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.HelmApplication{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))

	webservice.Route(webservice.GET("/categories").
		To(handler.ListCategories).
		Doc("List categories").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{v1alpha1.HelmCategory{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")))

	webservice.Route(webservice.GET("/categories/{category}").
		To(handler.DescribeCategory).
		Doc("Describe the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix.Category{}).
		Param(webservice.PathParameter("category", "category id")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}))

	c.Add(webservice)
	return nil
}
