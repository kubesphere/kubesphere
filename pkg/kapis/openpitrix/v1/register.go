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

package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	openpitrix2 "kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	op "kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"net/http"
)

const (
	GroupName = "openpitrix.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, op op.Client) error {
	if op == nil {
		return nil
	}
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}
	webservice := runtime.NewWebService(GroupVersion)
	handler := newOpenpitrixHandler(factory, op)

	webservice.Route(webservice.GET("/applications").
		To(handler.ListApplications).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/namespaces/{namespace}/applications").
		To(handler.ListApplications).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications within the specified namespace").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.PathParameter("namespace", "the name of the project.").Required(true)).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/namespaces/{namespace}/applications/{application}").
		To(handler.DescribeApplication).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Application{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.GET("/workspaces/{workspace}/clusters/{cluster}/applications").
		To(handler.ListApplications).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications in special cluster").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications").
		To(handler.ListApplications).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications within the specified namespace").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications/{application}").
		To(handler.DescribeApplication).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Application{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.POST("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications").
		To(handler.CreateApplication).
		Doc("Deploy a new application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Reads(openpitrix2.CreateClusterRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)))

	webservice.Route(webservice.PATCH("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications/{application}").
		Consumes(mimePatch...).
		To(handler.ModifyApplication).
		Doc("Modify application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Reads(openpitrix2.ModifyClusterAttributesRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.DELETE("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications/{application}").
		To(handler.DeleteApplication).
		Doc("Delete the specified application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.POST("/workspaces/{workspace}/clusters/{cluster}/namespaces/{namespace}/applications/{application}").
		Consumes(mimePatch...).
		To(handler.UpgradeApplication).
		Doc("Upgrade application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Reads(openpitrix2.UpgradeClusterRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("cluster", "the name of the cluster.").Required(true)).
		Param(webservice.PathParameter("namespace", "the name of the project").Required(true)).
		Param(webservice.PathParameter("application", "the id of the application").Required(true)))

	webservice.Route(webservice.POST("/apps/{app}/versions").
		To(handler.CreateAppVersion).
		Doc("Create a new app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.CreateAppVersionRequest{}).
		Param(webservice.QueryParameter("validate", "Validate format of package(pack by op tool)")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateAppVersionResponse{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/workspaces/{workspace}/apps/{app}/versions").
		To(handler.CreateAppVersion).
		Doc("Create a new app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.CreateAppVersionRequest{}).
		Param(webservice.QueryParameter("validate", "Validate format of package(pack by op tool)")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateAppVersionResponse{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/apps/{app}/versions/{version}").
		To(handler.DeleteAppVersion).
		Doc("Delete the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/workspaces/{workspace}/apps/{app}/versions/{version}").
		To(handler.DeleteAppVersion).
		Doc("Delete the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/apps/{app}/versions/{version}").
		Consumes(mimePatch...).
		To(handler.ModifyAppVersion).
		Doc("Patch the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/workspaces/{workspace}/apps/{app}/versions/{version}").
		Consumes(mimePatch...).
		To(handler.ModifyAppVersion).
		Doc("Patch the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}").
		To(handler.DescribeAppVersion).
		Doc("Describe the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersion{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions").
		To(handler.ListAppVersions).
		Doc("Get active versions of app, can filter with these fields(version_id, app_id, name, owner, description, package_name, status, type), default return all active app versions").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions,connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.PathParameter("app", "app template id")).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}))
	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}/versions/{version}").
		To(handler.DescribeAppVersion).
		Doc("Describe the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersion{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}/versions").
		To(handler.ListAppVersions).
		Doc("Get active versions of app, can filter with these fields(version_id, app_id, name, owner, description, package_name, status, type), default return all active app versions").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions,connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(webservice.PathParameter("app", "app template id")).
		Param(webservice.QueryParameter(params.ReverseParam, "sort parameters, e.g. reverse=true")).
		Param(webservice.QueryParameter(params.OrderByParam, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}))
	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}/versions/{version}/audits").
		To(handler.ListAppVersionAudits).
		Doc("List audits information of version-specific app template").
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersionAudit{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/audits").
		To(handler.ListAppVersionAudits).
		Doc("List audits information of version-specific app template").
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersionAudit{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/package").
		To(handler.GetAppVersionPackage).
		Doc("Get packages of version-specific app").
		Returns(http.StatusOK, api.StatusOK, openpitrix2.GetAppVersionPackageResponse{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/apps/{app}/versions/{version}/action").
		To(handler.DoAppVersionAction).
		Doc("Perform submit or other operations on app").
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/workspaces/{workspace}/apps/{app}/versions/{version}/action").
		To(handler.DoAppVersionAction).
		Doc("Perform submit or other operations on app").
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/files").
		To(handler.GetAppVersionFiles).
		Doc("Get app template package files").
		Returns(http.StatusOK, api.StatusOK, openpitrix2.GetAppVersionPackageFilesResponse{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/reviews").
		To(handler.ListReviews).
		Doc("Get reviews of version-specific app").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions,connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersionReview{}))
	webservice.Route(webservice.GET("/apps/{app}/audits").
		To(handler.ListAppVersionAudits).
		Doc("List audits information of the specific app template").
		Param(webservice.PathParameter("app", "app template id")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersionAudit{}))
	webservice.Route(webservice.POST("/apps").
		To(handler.CreateApp).
		Doc("Create a new app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateAppResponse{}).
		Reads(openpitrix2.CreateAppRequest{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/workspaces/{workspace}/apps").
		To(handler.CreateApp).
		Doc("Create a new app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateAppResponse{}).
		Reads(openpitrix2.CreateAppRequest{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/apps/{app}").
		To(handler.DeleteApp).
		Doc("Delete the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/workspaces/{workspace}/apps/{app}").
		To(handler.DeleteApp).
		Doc("Delete the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/apps/{app}").
		Consumes(mimePatch...).
		To(handler.ModifyApp).
		Doc("Patch the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/workspaces/{workspace}/apps/{app}").
		Consumes(mimePatch...).
		To(handler.ModifyApp).
		Doc("Patch the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/workspaces/{workspace}/apps/{app}").
		To(handler.DescribeApp).
		Doc("Describe the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersion{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}").
		To(handler.DescribeApp).
		Doc("Describe the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.AppVersion{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/apps/{app}/action").
		To(handler.DoAppAction).
		Doc("Perform recover or suspend operation on app").
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/workspaces/{workspace}/apps/{app}/action").
		To(handler.DoAppAction).
		Doc("Perform recover or suspend operation on app").
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps").
		To(handler.ListApps).
		Doc("List app templates").
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
	webservice.Route(webservice.GET("/workspaces/{workspace}/apps").
		To(handler.ListApps).
		Doc("List app templates in the specified workspace.").
		Param(webservice.PathParameter("workspace", "workspace name")).
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
	webservice.Route(webservice.POST("/categories").
		To(handler.CreateCategory).
		Doc("Create app template category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.CreateCategoryRequest{}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateCategoryResponse{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/categories/{category}").
		To(handler.DeleteCategory).
		Doc("Delete the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.PATCH("/categories/{category}").
		Consumes(mimePatch...).
		To(handler.ModifyCategory).
		Doc("Patch the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyCategoryRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.GET("/categories/{category}").
		To(handler.DescribeCategory).
		Doc("Describe the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Category{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.GET("/categories").
		To(handler.ListCategories).
		Doc("List categories").
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

	webservice.Route(webservice.GET("/attachments/{attachment}").
		To(handler.DescribeAttachment).
		Doc("Get attachment by attachment id").
		Param(webservice.PathParameter("attachment", "attachment id")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Attachment{}))

	webservice.Route(webservice.POST("/repos").
		To(handler.CreateRepo).
		Doc("Create repository in the specified workspace, repository used to store package of app").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Param(webservice.QueryParameter("validate", "Validate repository")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateRepoResponse{}).
		Reads(openpitrix2.CreateRepoRequest{}))
	webservice.Route(webservice.POST("/workspaces/{workspace}/repos").
		To(handler.CreateRepo).
		Doc("Create repository in the specified workspace, repository used to store package of app").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Param(webservice.QueryParameter("validate", "Validate repository")).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.CreateRepoResponse{}).
		Reads(openpitrix2.CreateRepoRequest{}))
	webservice.Route(webservice.DELETE("/repos/{repo}").
		To(handler.DeleteRepo).
		Doc("Delete the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.DELETE("/workspaces/{workspace}/repos/{repo}").
		To(handler.DeleteRepo).
		Doc("Delete the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.PATCH("/repos/{repo}").
		Consumes(mimePatch...).
		To(handler.ModifyRepo).
		Doc("Patch the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyRepoRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.PATCH("/workspaces/{workspace}/repos/{repo}").
		Consumes(mimePatch...).
		To(handler.ModifyRepo).
		Doc("Patch the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(openpitrix2.ModifyRepoRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos/{repo}").
		To(handler.DescribeRepo).
		Doc("Describe the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Repo{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/workspaces/{workspace}/repos/{repo}").
		To(handler.DescribeRepo).
		Doc("Describe the specified repository in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, api.StatusOK, openpitrix2.Repo{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos").
		To(handler.ListRepos).
		Doc("List repositories in the specified workspace").
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
	webservice.Route(webservice.GET("/workspaces/{workspace}/repos").
		To(handler.ListRepos).
		Doc("List repositories in the specified workspace").
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

	webservice.Route(webservice.POST("/repos/{repo}/action").
		To(handler.DoRepoAction).
		Doc("Start index repository event").
		Reads(openpitrix2.RepoActionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.POST("/workspaces/{workspace}/repos/{repo}/action").
		To(handler.DoRepoAction).
		Doc("Start index repository event").
		Reads(openpitrix2.RepoActionRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos/{repo}/events").
		To(handler.ListRepoEvents).
		Doc("Get repository events").
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/workspaces/{workspace}/repos/{repo}/events").
		To(handler.ListRepoEvents).
		Doc("Get repository events").
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Param(webservice.PathParameter("repo", "repo id")))

	c.Add(webservice)

	return nil
}
