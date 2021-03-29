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
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	openpitrixoptions "kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"net/url"
	"strconv"
	"strings"
)

type openpitrixHandler struct {
	openpitrix openpitrix.Interface
}

func newOpenpitrixHandler(ksInformers informers.InformerFactory, ksClient versioned.Interface, option *openpitrixoptions.Options) *openpitrixHandler {
	var s3Client s3.Interface
	if option != nil && option.S3Options != nil && len(option.S3Options.Endpoint) != 0 {
		var err error
		s3Client, err = s3.NewS3Client(option.S3Options)
		if err != nil {
			klog.Errorf("failed to connect to storage, please check storage service status, error: %v", err)
		}
	}

	return &openpitrixHandler{
		openpitrix.NewOpenpitrixOperator(ksInformers, ksClient, s3Client),
	}
}

func (h *openpitrixHandler) CreateRepo(req *restful.Request, resp *restful.Response) {

	createRepoRequest := &openpitrix.CreateRepoRequest{}
	err := req.ReadEntity(createRepoRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	createRepoRequest.Workspace = new(string)
	*createRepoRequest.Workspace = req.PathParameter("workspace")

	user, _ := request.UserFrom(req.Request.Context())
	creator := ""
	if user != nil {
		creator = user.GetName()
	}
	parsedUrl, err := url.Parse(createRepoRequest.URL)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	userInfo := parsedUrl.User
	// trim credential from url
	parsedUrl.User = nil

	repo := v1alpha1.HelmRepo{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmRepoIdPrefix),
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: creator,
			},
			Labels: map[string]string{
				constants.WorkspaceLabelKey: *createRepoRequest.Workspace,
			},
		},
		Spec: v1alpha1.HelmRepoSpec{
			Name:        createRepoRequest.Name,
			Url:         parsedUrl.String(),
			SyncPeriod:  0,
			Description: stringutils.ShortenString(createRepoRequest.Description, 512),
		},
	}

	if strings.HasPrefix(createRepoRequest.URL, "https://") || strings.HasPrefix(createRepoRequest.URL, "http://") {
		if userInfo != nil {
			repo.Spec.Credential.Username = userInfo.Username()
			repo.Spec.Credential.Password, _ = userInfo.Password()
		}
	} else if strings.HasPrefix(createRepoRequest.URL, "s3://") {
		cfg := v1alpha1.S3Config{}
		err := json.Unmarshal([]byte(createRepoRequest.Credential), &cfg)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
		repo.Spec.Credential.S3Config = cfg
	}

	var result interface{}
	// 1. validate repo
	result, err = h.openpitrix.ValidateRepo(createRepoRequest.URL, &repo.Spec.Credential)
	if err != nil {
		klog.Errorf("validate repo failed, err: %s", err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	// 2. create repo
	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))
	if !validate {
		if repo.GetTrueName() == "" {
			api.HandleBadRequest(resp, nil, fmt.Errorf("repo name is empty"))
			return
		}
		result, err = h.openpitrix.CreateRepo(&repo)
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
	repoActionRequest.Workspace = req.PathParameter("workspace")

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
	} else {
		klog.V(4).Info("delete repo: ", repoId)
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
	} else {
		klog.V(4).Info("modify repo: ", repoId)
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
	if status.Code(err) == codes.InvalidArgument || status.Code(err) == codes.FailedPrecondition {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	klog.Errorln(err)
	api.HandleInternalError(resp, nil, err)
}

//=============
// helm application template handler
//=============

func (h *openpitrixHandler) DoAppAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest openpitrix.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		doActionRequest.Username = user.GetName()
	}

	appId := strings.TrimSuffix(req.PathParameter("app"), v1alpha1.HelmApplicationAppStoreSuffix)

	err = h.openpitrix.DoAppAction(appId, &doActionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
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

	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		createAppRequest.Username = user.GetName()
	}

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
		_ = validatePackageRequest
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

func (h *openpitrixHandler) ListApps(req *restful.Request, resp *restful.Response) {
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

	if conditions.Match[openpitrix.WorkspaceLabel] == "" {
		conditions.Match[openpitrix.WorkspaceLabel] = req.QueryParameter("workspace")
	}

	result, err := h.openpitrix.ListApps(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(result)
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
	appId := strings.TrimSuffix(req.PathParameter("app"), v1alpha1.HelmApplicationAppStoreSuffix)

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

// app version
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
	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		createAppVersionRequest.Username = user.GetName()
	}
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

func (h *openpitrixHandler) ListAppVersions(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	appId := req.PathParameter("app")
	appId = strings.TrimSuffix(appId, v1alpha1.HelmApplicationAppStoreSuffix)

	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	conditions.Match[openpitrix.AppId] = appId

	if req.PathParameter("workspace") != "" {
		conditions.Match[openpitrix.WorkspaceLabel] = req.PathParameter("workspace")
	}

	if conditions.Match[openpitrix.WorkspaceLabel] == "" {
		conditions.Match[openpitrix.WorkspaceLabel] = req.QueryParameter("workspace")
	}

	result, err := h.openpitrix.ListAppVersions(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
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

func (h *openpitrixHandler) GetAppVersionFiles(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	getAppVersionFilesRequest := &openpitrix.GetAppVersionFilesRequest{}
	if f := req.QueryParameter("files"); f != "" {
		getAppVersionFilesRequest.Files = strings.Split(f, ",")
	}

	result, err := h.openpitrix.GetAppVersionFiles(versionId, getAppVersionFilesRequest)

	if err != nil {
		klog.Errorln(err)
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, nil, err)
		} else {
			api.HandleBadRequest(resp, nil, err)
		}
		return
	}

	resp.WriteEntity(result)
}

// app version audit
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

	conditions.Match[openpitrix.AppId] = strings.TrimSuffix(appId, v1alpha1.HelmApplicationAppStoreSuffix)
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

func (h *openpitrixHandler) DoAppVersionAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest openpitrix.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	doActionRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		doActionRequest.Username = user.GetName()
	}
	versionId := req.PathParameter("version")

	err = h.openpitrix.DoAppVersionAction(versionId, &doActionRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

// application release

func (h *openpitrixHandler) DescribeApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	workspace := req.PathParameter("workspace")
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")

	app, err := h.openpitrix.DescribeApplication(workspace, clusterName, namespace, applicationId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(app)
	return
}

func (h *openpitrixHandler) DeleteApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	workspace := req.PathParameter("workspace")
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")

	err := h.openpitrix.DeleteApplication(workspace, clusterName, namespace, applicationId)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) ListApplications(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("workspace")
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, openpitrix.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	conditions, err := params.ParseConditions(req)

	if offset < 0 {
		offset = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.openpitrix.ListApplications(workspace, clusterName, namespace, conditions, limit, offset, orderBy, reverse)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *openpitrixHandler) UpgradeApplication(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	applicationId := req.PathParameter("application")
	var upgradeClusterRequest openpitrix.UpgradeClusterRequest
	err := req.ReadEntity(&upgradeClusterRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	upgradeClusterRequest.Namespace = namespace
	upgradeClusterRequest.ClusterId = applicationId
	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		upgradeClusterRequest.Username = user.GetName()
	}

	err = h.openpitrix.UpgradeApplication(upgradeClusterRequest)
	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) ModifyApplication(req *restful.Request, resp *restful.Response) {
	var modifyClusterAttributesRequest openpitrix.ModifyClusterAttributesRequest
	applicationId := req.PathParameter("application")
	namespace := req.PathParameter("namespace")
	err := req.ReadEntity(&modifyClusterAttributesRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	modifyClusterAttributesRequest.Namespace = namespace
	modifyClusterAttributesRequest.ClusterID = applicationId

	err = h.openpitrix.ModifyApplication(modifyClusterAttributesRequest)

	if err != nil {
		klog.Errorln(err)
		handleOpenpitrixError(resp, err)
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *openpitrixHandler) CreateApplication(req *restful.Request, resp *restful.Response) {
	clusterName := req.PathParameter("cluster")
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("workspace")
	var createClusterRequest openpitrix.CreateClusterRequest
	err := req.ReadEntity(&createClusterRequest)
	createClusterRequest.Workspace = workspace
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		createClusterRequest.Username = user.GetName()
	}

	err = h.openpitrix.CreateApplication(workspace, clusterName, namespace, createClusterRequest)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(errors.None)
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

	resp.WriteEntity(result)
}

// review
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

func (h *openpitrixHandler) CreateAttachment(req *restful.Request, resp *restful.Response) {

	r := req.Request
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	var att *openpitrix.Attachment
	// just save one attachment
	for fName := range r.MultipartForm.File {
		f, _, err := r.FormFile(fName)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
		data, err := ioutil.ReadAll(f)
		f.Close()

		att, err = h.openpitrix.CreateAttachment(data)
		if err != nil {
			api.HandleInternalError(resp, nil, err)
			return
		}
		break
	}

	resp.WriteEntity(att)
}

func (h *openpitrixHandler) DeleteAttachments(req *restful.Request, resp *restful.Response) {
	attachmentId := req.PathParameter("attachment")

	ids := strings.Split(attachmentId, ",")
	err := h.openpitrix.DeleteAttachments(ids)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
	}
}
