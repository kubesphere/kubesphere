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
	"fmt"
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	op "kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"strconv"
	"strings"
)

type openpitrixHandler struct {
	openpitrix openpitrix.Interface
	informers  k8sinformers.SharedInformerFactory
}

func newOpenpitrixHandler(factory informers.InformerFactory, opClient op.Client) *openpitrixHandler {

	return &openpitrixHandler{
		openpitrix: openpitrix.NewOpenpitrixOperator(factory.KubernetesSharedInformerFactory(), opClient),
		informers:  factory.KubernetesSharedInformerFactory(),
	}
}

func (h *openpitrixHandler) ListApplications(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	conditions.Match[openpitrix.Zone] = namespace
	// in openpitrix, runtime id is the cluster name
	conditions.Match[openpitrix.RuntimeId] = clusterName

	result, err := h.openpitrix.ListApplications(conditions, limit, offset, orderBy, reverse)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) DescribeApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	applicationId := req.PathParameter("application")

	app, err := h.openpitrix.DescribeApplication(namespace, applicationId, clusterName)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(app)
	return
}

func (h *openpitrixHandler) CreateApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	var createClusterRequest openpitrix.CreateClusterRequest
	err := req.ReadEntity(&createClusterRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	createClusterRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	err = h.openpitrix.CreateApplication(clusterName, namespace, createClusterRequest)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) ModifyApplication(req *restful.Request, resp *restful.Response) {
	var modifyClusterAttributesRequest openpitrix.ModifyClusterAttributesRequest
	clusterName := req.PathParameter("cluster")
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")
	err := req.ReadEntity(&modifyClusterAttributesRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	app, err := h.openpitrix.DescribeApplication(namespace, applicationId, clusterName)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	if clusterName != app.Cluster.RuntimeId {
		err = fmt.Errorf("runtime and cluster not match %s,%s", app.Cluster.RuntimeId, clusterName)
		klog.V(4).Infoln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	err = h.openpitrix.ModifyApplication(modifyClusterAttributesRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DeleteApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")
	app, err := h.openpitrix.DescribeApplication(namespace, applicationId, clusterName)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	if clusterName != app.Cluster.RuntimeId {
		err = fmt.Errorf("runtime and cluster not match %s,%s", app.Cluster.RuntimeId, clusterName)
		klog.V(4).Infoln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	err = h.openpitrix.DeleteApplication(applicationId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) UpgradeApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	applicationId := req.PathParameter("application")
	var upgradeClusterRequest openpitrix.UpgradeClusterRequest
	err := req.ReadEntity(&upgradeClusterRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	upgradeClusterRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	app, err := h.openpitrix.DescribeApplication(namespace, applicationId, clusterName)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	if clusterName != app.Cluster.RuntimeId {
		err = fmt.Errorf("runtime and cluster not match %s,%s", app.Cluster.RuntimeId, clusterName)
		klog.V(4).Infoln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	err = h.openpitrix.UpgradeApplication(upgradeClusterRequest)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) GetAppVersionPackage(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")
	versionId := req.PathParameter("version")

	result, err := h.openpitrix.GetAppVersionPackage(appId, versionId)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) DoAppAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest openpitrix.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	appId := req.PathParameter("app")

	err = h.openpitrix.DoAppAction(appId, &doActionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DoAppVersionAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest openpitrix.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	doActionRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	versionId := req.PathParameter("version")

	err = h.openpitrix.DoAppVersionAction(versionId, &doActionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) GetAppVersionFiles(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	getAppVersionFilesRequest := &openpitrix.GetAppVersionFilesRequest{}
	if f := req.QueryParameter("files"); f != "" {
		getAppVersionFilesRequest.Files = strings.Split(f, ",")
	}

	result, err := h.openpitrix.GetAppVersionFiles(versionId, getAppVersionFilesRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListAppVersionAudits(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.StatusTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	appId := req.PathParameter("app")
	versionId := req.PathParameter("version")
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	conditions.Match[openpitrix.AppId] = appId
	if versionId != "" {
		conditions.Match[openpitrix.VersionId] = versionId
	}

	result, err := h.openpitrix.ListAppVersionAudits(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListReviews(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.StatusTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.ListAppVersionReviews(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListAppVersions(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	appId := req.PathParameter("app")
	statistics := params.GetBoolValueWithDefault(req, "statistics", false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	conditions.Match[openpitrix.AppId] = appId

	result, err := h.openpitrix.ListAppVersions(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	if statistics {
		for _, item := range result.Items {
			if version, ok := item.(*openpitrix.AppVersion); ok {
				statisticsResult, err := h.openpitrix.ListApplications(&params.Conditions{Match: map[string]string{openpitrix.AppId: version.AppId, openpitrix.VersionId: version.VersionId}}, 0, 0, "", false)
				if err != nil {
					klog.Errorln(err)
					api.HandleInternalError(resp, nil, err)
					return
				}
				version.ClusterTotal = &statisticsResult.TotalCount
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListApps(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	statistics := params.GetBoolValueWithDefault(req, "statistics", false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	if req.PathParameter("workspace") != "" {
		conditions.Match[openpitrix.ISV] = req.PathParameter("workspace")
	}

	if conditions.Match[openpitrix.ISV] == "" {
		conditions.Match[openpitrix.ISV] = req.QueryParameter("workspace")
	}

	result, err := h.openpitrix.ListApps(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	if statistics {
		for _, item := range result.Items {
			if app, ok := item.(*openpitrix.App); ok {
				statuses := "active|used|enabled|stopped|pending|creating|upgrading|updating|rollbacking|stopping|starting|recovering|resizing|scaling|deleting"
				statisticsResult, err := h.openpitrix.ListApplications(&params.Conditions{Match: map[string]string{openpitrix.AppId: app.AppId, openpitrix.Status: statuses}}, 0, 0, "", false)
				if err != nil {
					klog.Errorln(err)
					handleOpenpitrixError(resp, err)
					return
				}
				app.ClusterTotal = &statisticsResult.TotalCount
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ModifyApp(req *restful.Request, resp *restful.Response) {

	var patchAppRequest openpitrix.ModifyAppRequest
	err := req.ReadEntity(&patchAppRequest)
	appId := req.PathParameter("app")

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	err = h.openpitrix.ModifyApp(appId, &patchAppRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DescribeApp(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")

	result, err := h.openpitrix.DescribeApp(appId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) DeleteApp(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")

	err := h.openpitrix.DeleteApp(appId)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			api.HandleNotFound(resp, nil, err)
			return
		}
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) CreateApp(req *restful.Request, resp *restful.Response) {
	createAppRequest := &openpitrix.CreateAppRequest{}
	err := req.ReadEntity(createAppRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	createAppRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	if req.PathParameter("workspace") != "" {
		createAppRequest.Isv = req.PathParameter("workspace")
	}

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validatePackageRequest := &openpitrix.ValidatePackageRequest{
			VersionPackage: createAppRequest.VersionPackage,
			VersionType:    createAppRequest.VersionType,
		}
		result, err = h.openpitrix.ValidatePackage(validatePackageRequest)
	} else {
		result, err = h.openpitrix.CreateApp(createAppRequest)
	}

	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			api.HandleBadRequest(resp, nil, err)
			return
		}
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) CreateAppVersion(req *restful.Request, resp *restful.Response) {
	var createAppVersionRequest openpitrix.CreateAppVersionRequest
	err := req.ReadEntity(&createAppVersionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	// override app id
	createAppVersionRequest.AppId = req.PathParameter("app")
	createAppVersionRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validatePackageRequest := &openpitrix.ValidatePackageRequest{
			VersionPackage: createAppVersionRequest.Package,
			VersionType:    createAppVersionRequest.Type,
		}
		result, err = h.openpitrix.ValidatePackage(validatePackageRequest)
	} else {
		result, err = h.openpitrix.CreateAppVersion(&createAppVersionRequest)
	}

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ModifyAppVersion(req *restful.Request, resp *restful.Response) {

	var patchAppVersionRequest openpitrix.ModifyAppVersionRequest
	err := req.ReadEntity(&patchAppVersionRequest)
	versionId := req.PathParameter("version")

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	err = h.openpitrix.ModifyAppVersion(versionId, &patchAppVersionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DeleteAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")

	err := h.openpitrix.DeleteAppVersion(versionId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DescribeAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")

	result, err := h.openpitrix.DescribeAppVersion(versionId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) DescribeAttachment(req *restful.Request, resp *restful.Response) {
	attachmentId := req.PathParameter("attachment")
	fileName := req.QueryParameter("filename")
	result, err := h.openpitrix.DescribeAttachment(attachmentId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	// file raw
	if fileName != "" {
		data := result.AttachmentContent[fileName]
		resp.Write(data)
		resp.Header().Set("Content-Type", "text/plain")
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) CreateCategory(req *restful.Request, resp *restful.Response) {
	createCategoryRequest := &openpitrix.CreateCategoryRequest{}
	err := req.ReadEntity(createCategoryRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.CreateCategory(createCategoryRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}
func (h *openpitrixHandler) DeleteCategory(req *restful.Request, resp *restful.Response) {
	categoryId := req.PathParameter("category")

	err := h.openpitrix.DeleteCategory(categoryId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}
func (h *openpitrixHandler) ModifyCategory(req *restful.Request, resp *restful.Response) {
	var modifyCategoryRequest openpitrix.ModifyCategoryRequest
	categoryId := req.PathParameter("category")
	err := req.ReadEntity(&modifyCategoryRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	err = h.openpitrix.ModifyCategory(categoryId, &modifyCategoryRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}
func (h *openpitrixHandler) DescribeCategory(req *restful.Request, resp *restful.Response) {
	categoryId := req.PathParameter("category")

	result, err := h.openpitrix.DescribeCategory(categoryId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}
func (h *openpitrixHandler) ListCategories(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	statistics := params.GetBoolValueWithDefault(req, "statistics", false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.ListCategories(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	if statistics {
		for _, item := range result.Items {
			if category, ok := item.(*openpitrix.Category); ok {
				statisticsResult, err := h.openpitrix.ListApps(&params.Conditions{
					Match: map[string]string{
						openpitrix.CategoryId: category.CategoryID,
						openpitrix.Status:     openpitrix.StatusActive,
						openpitrix.RepoId:     openpitrix.BuiltinRepoId,
					},
				}, "", false, 0, 0)
				if err != nil {
					klog.Errorln(err)
					handleOpenpitrixError(resp, err)
					return
				}
				category.AppTotal = &statisticsResult.TotalCount
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) CreateRepo(req *restful.Request, resp *restful.Response) {
	createRepoRequest := &openpitrix.CreateRepoRequest{}
	err := req.ReadEntity(createRepoRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	if req.PathParameter("workspace") != "" {
		if createRepoRequest.Workspace == nil {
			createRepoRequest.Workspace = new(string)
		}
		*createRepoRequest.Workspace = req.PathParameter("workspace")
	}

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validateRepoRequest := &openpitrix.ValidateRepoRequest{
			Type:       createRepoRequest.Type,
			Url:        createRepoRequest.URL,
			Credential: createRepoRequest.Credential,
		}
		result, err = h.openpitrix.ValidateRepo(validateRepoRequest)
	} else {
		result, err = h.openpitrix.CreateRepo(createRepoRequest)
	}

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) DoRepoAction(req *restful.Request, resp *restful.Response) {
	repoActionRequest := &openpitrix.RepoActionRequest{}
	repoId := req.PathParameter("repo")
	err := req.ReadEntity(repoActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	err = h.openpitrix.DoRepoAction(repoId, repoActionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DeleteRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	err := h.openpitrix.DeleteRepo(repoId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) ModifyRepo(req *restful.Request, resp *restful.Response) {
	var updateRepoRequest openpitrix.ModifyRepoRequest
	repoId := req.PathParameter("repo")
	err := req.ReadEntity(&updateRepoRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	err = h.openpitrix.ModifyRepo(repoId, &updateRepoRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) DescribeRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	result, err := h.openpitrix.DescribeRepo(repoId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListRepos(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	if req.PathParameter("workspace") != "" {
		conditions.Match[openpitrix.WorkspaceLabel] = req.PathParameter("workspace")
	}

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.ListRepos(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *openpitrixHandler) ListRepoEvents(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")
	limit, offset := params.ParsePaging(req)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.ListRepoEvents(repoId, conditions, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func handleOpenpitrixError(resp *restful.Response, err error) {
	if status.Code(err) == codes.NotFound {
		klog.V(4).Infoln(err)
		api.HandleNotFound(resp, nil, err)
		return
	}
	if status.Code(err) == codes.InvalidArgument {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	klog.Errorln(err)
	api.HandleInternalError(resp, nil, err)
}
