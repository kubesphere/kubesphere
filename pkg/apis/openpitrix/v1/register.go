/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */
package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/openpitrix"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	opmodels "kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

const GroupName = "openpitrix.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	ok := "ok"
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}
	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.GET("/applications").
		To(openpitrix.ListApplications).
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/namespaces/{namespace}/applications").
		To(openpitrix.ListApplications).
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("List all applications within the specified namespace").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions, connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=value,key~value").
			DefaultValue("")).
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")))

	webservice.Route(webservice.GET("/namespaces/{namespace}/applications/{application}").
		To(openpitrix.DescribeApplication).
		Returns(http.StatusOK, ok, opmodels.Application{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Doc("Describe the specified application of the namespace").
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.PathParameter("application", "application ID")))

	webservice.Route(webservice.POST("/namespaces/{namespace}/applications").
		To(openpitrix.CreateApplication).
		Doc("Deploy a new application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Reads(opmodels.CreateClusterRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("namespace", "the name of the project")))

	webservice.Route(webservice.PATCH("/namespaces/{namespace}/applications/{application}").
		Consumes(mimePatch...).
		To(openpitrix.ModifyApplication).
		Doc("Modify application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Reads(opmodels.ModifyClusterAttributesRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.PathParameter("application", "the id of the application cluster")))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/applications/{application}").
		To(openpitrix.DeleteApplication).
		Doc("Delete the specified application").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceResourcesTag}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Param(webservice.PathParameter("application", "the id of the application cluster")))

	webservice.Route(webservice.POST("/apps/{app}/versions").
		To(openpitrix.CreateAppVersion).
		Doc("Create a new app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.CreateAppVersionRequest{}).
		Param(webservice.QueryParameter("validate", "Validate format of package(pack by op tool)")).
		Returns(http.StatusOK, ok, opmodels.CreateAppVersionResponse{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/apps/{app}/versions/{version}").
		To(openpitrix.DeleteAppVersion).
		Doc("Delete the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/apps/{app}/versions/{version}").
		Consumes(mimePatch...).
		To(openpitrix.ModifyAppVersion).
		Doc("Patch the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}").
		To(openpitrix.DescribeAppVersion).
		Doc("Describe the specified app template version").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, opmodels.AppVersion{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions").
		To(openpitrix.ListAppVersions).
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
		Returns(http.StatusOK, ok, models.PageableResponse{}))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/audits").
		To(openpitrix.ListAppVersionAudits).
		Doc("List audits information of version-specific app template").
		Returns(http.StatusOK, ok, opmodels.AppVersionAudit{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/package").
		To(openpitrix.GetAppVersionPackage).
		Doc("Get packages of version-specific app").
		Returns(http.StatusOK, ok, opmodels.GetAppVersionPackageResponse{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/apps/{app}/versions/{version}/action").
		To(openpitrix.DoAppVersionAction).
		Doc("Perform submit or other operations on app").
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}/versions/{version}/files").
		To(openpitrix.GetAppVersionFiles).
		Doc("Get app template package files").
		Returns(http.StatusOK, ok, opmodels.GetAppVersionPackageFilesResponse{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/reviews").
		To(openpitrix.ListReviews).
		Doc("Get reviews of version-specific app").
		Param(webservice.QueryParameter(params.ConditionsParam, "query conditions,connect multiple conditions with commas, equal symbol for exact query, wave symbol for fuzzy query e.g. name~a").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Param(webservice.QueryParameter(params.PagingParam, "paging query, e.g. limit=100,page=1").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Returns(http.StatusOK, ok, opmodels.AppVersionReview{}))
	webservice.Route(webservice.GET("/apps/{app}/audits").
		To(openpitrix.ListAppVersionAudits).
		Doc("List audits information of the specific app template").
		Param(webservice.PathParameter("app", "app template id")).
		Returns(http.StatusOK, ok, opmodels.AppVersionAudit{}))
	webservice.Route(webservice.POST("/apps").
		To(openpitrix.CreateApp).
		Doc("Create a new app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, opmodels.CreateAppResponse{}).
		Reads(opmodels.CreateAppRequest{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/apps/{app}").
		To(openpitrix.DeleteApp).
		Doc("Delete the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.PATCH("/apps/{app}").
		Consumes(mimePatch...).
		To(openpitrix.ModifyApp).
		Doc("Patch the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.ModifyAppVersionRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps/{app}").
		To(openpitrix.DescribeApp).
		Doc("Describe the specified app template").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, opmodels.AppVersion{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.POST("/apps/{app}/action").
		To(openpitrix.DoAppAction).
		Doc("Perform recover or suspend operation on app").
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("version", "app template version id")).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.GET("/apps").
		To(openpitrix.ListApps).
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
		Returns(http.StatusOK, ok, models.PageableResponse{}))
	webservice.Route(webservice.POST("/categories").
		To(openpitrix.CreateCategory).
		Doc("Create app template category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.CreateCategoryRequest{}).
		Returns(http.StatusOK, ok, opmodels.CreateCategoryResponse{}).
		Param(webservice.PathParameter("app", "app template id")))
	webservice.Route(webservice.DELETE("/categories/{category}").
		To(openpitrix.DeleteCategory).
		Doc("Delete the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.PATCH("/categories/{category}").
		Consumes(mimePatch...).
		To(openpitrix.ModifyCategory).
		Doc("Patch the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.ModifyCategoryRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.GET("/categories/{category}").
		To(openpitrix.DescribeCategory).
		Doc("Describe the specified category").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, opmodels.Category{}).
		Param(webservice.PathParameter("category", "category id")))
	webservice.Route(webservice.GET("/categories").
		To(openpitrix.ListCategories).
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
		Returns(http.StatusOK, ok, models.PageableResponse{}))

	webservice.Route(webservice.GET("/attachments/{attachment}").
		To(openpitrix.DescribeAttachment).
		Doc("Get attachment by attachment id").
		Param(webservice.PathParameter("attachment", "attachment id")).
		Returns(http.StatusOK, ok, opmodels.Attachment{}))

	webservice.Route(webservice.POST("/repos").
		To(openpitrix.CreateRepo).
		Doc("Create repository, repository used to store package of app").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Param(webservice.QueryParameter("validate", "Validate repository")).
		Returns(http.StatusOK, ok, opmodels.CreateRepoResponse{}).
		Reads(opmodels.CreateRepoRequest{}))
	webservice.Route(webservice.DELETE("/repos/{repo}").
		To(openpitrix.DeleteRepo).
		Doc("Delete the specified repository").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.PATCH("/repos/{repo}").
		Consumes(mimePatch...).
		To(openpitrix.ModifyRepo).
		Doc("Patch the specified repository").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Reads(opmodels.ModifyRepoRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos/{repo}").
		To(openpitrix.DescribeRepo).
		Doc("Describe the specified repository").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.OpenpitrixTag}).
		Returns(http.StatusOK, ok, opmodels.Repo{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos").
		To(openpitrix.ListRepos).
		Doc("List repositories").
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
	webservice.Route(webservice.POST("/repos/{repo}/action").
		To(openpitrix.DoRepoAction).
		Doc("Start index repository event").
		Reads(opmodels.RepoActionRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Param(webservice.PathParameter("repo", "repo id")))
	webservice.Route(webservice.GET("/repos/{repo}/events").
		To(openpitrix.ListRepoEvents).
		Doc("Get repository events").
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Param(webservice.PathParameter("repo", "repo id")))

	c.Add(webservice)

	return nil
}
