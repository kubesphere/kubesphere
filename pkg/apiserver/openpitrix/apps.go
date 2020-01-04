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

package openpitrix

import (
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/openpitrix/app"
	"kubesphere.io/kubesphere/pkg/models/openpitrix/type"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
	"strconv"
	"strings"
)

func GetAppVersionPackage(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")
	versionId := req.PathParameter("version")

	result, err := app.GetAppVersionPackage(appId, versionId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func DoAppAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest types.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	appId := req.PathParameter("app")

	err = app.DoAppAction(appId, &doActionRequest)
	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DoAppVersionAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest types.ActionRequest
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	doActionRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	versionId := req.PathParameter("version")

	err = app.DoAppVersionAction(versionId, &doActionRequest)
	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func GetAppVersionFiles(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	getAppVersionFilesRequest := &types.GetAppVersionFilesRequest{}
	if f := req.QueryParameter("files"); f != "" {
		getAppVersionFilesRequest.Files = strings.Split(f, ",")
	}

	result, err := app.GetAppVersionFiles(versionId, getAppVersionFilesRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func ListAppVersionAudits(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)
	appId := req.PathParameter("app")
	versionId := req.PathParameter("version")
	if orderBy == "" {
		orderBy = "status_time"
		reverse = true
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	conditions.Match["app"] = appId
	if versionId != "" {
		conditions.Match["version"] = versionId
	}

	result, err := app.ListAppVersionAudits(conditions, orderBy, reverse, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func ListReviews(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)
	if orderBy == "" {
		orderBy = "status_time"
		reverse = true
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := app.ListAppVersionReviews(conditions, orderBy, reverse, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func ListAppVersions(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)
	appId := req.PathParameter("app")
	statistics, _ := strconv.ParseBool(req.QueryParameter("statistics"))
	if orderBy == "" {
		orderBy = "create_time"
		reverse = true
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	conditions.Match["app"] = appId

	result, err := app.ListAppVersions(conditions, orderBy, reverse, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if statistics {
		for _, item := range result.Items {
			if version, ok := item.(*types.AppVersion); ok {
				statisticsResult, err := openpitrix.ListApplications(&params.Conditions{Match: map[string]string{"app_id": version.AppId, "version_id": version.VersionId}}, 0, 0, "", false)
				if err != nil {
					klog.Errorln(err)
					resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
					return
				}
				version.ClusterTotal = &statisticsResult.TotalCount
			}
		}
	}

	resp.WriteEntity(result)
}

func ListApps(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)
	statistics, _ := strconv.ParseBool(req.QueryParameter("statistics"))
	if orderBy == "" {
		orderBy = "create_time"
		reverse = true
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := app.ListApps(conditions, orderBy, reverse, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if statistics {
		for _, item := range result.Items {
			if app, ok := item.(*types.App); ok {
				status := "active|used|enabled|stopped|pending|creating|upgrading|updating|rollbacking|stopping|starting|recovering|resizing|scaling|deleting"
				statisticsResult, err := openpitrix.ListApplications(&params.Conditions{Match: map[string]string{"app_id": app.AppId, "status": status}}, 0, 0, "", false)
				if err != nil {
					klog.Errorln(err)
					resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
					return
				}
				app.ClusterTotal = &statisticsResult.TotalCount
			}
		}
	}

	resp.WriteEntity(result)
}

func ModifyApp(req *restful.Request, resp *restful.Response) {

	var patchAppRequest types.ModifyAppRequest
	err := req.ReadEntity(&patchAppRequest)
	appId := req.PathParameter("app")

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = app.PatchApp(appId, &patchAppRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DescribeApp(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")

	result, err := app.DescribeApp(appId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func DeleteApp(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")

	err := app.DeleteApp(appId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func CreateApp(req *restful.Request, resp *restful.Response) {
	createAppRequest := &types.CreateAppRequest{}
	err := req.ReadEntity(createAppRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	createAppRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validatePackageRequest := &types.ValidatePackageRequest{
			VersionPackage: createAppRequest.VersionPackage,
			VersionType:    createAppRequest.VersionType,
		}
		result, err = app.ValidatePackage(validatePackageRequest)
	} else {
		result, err = app.CreateApp(createAppRequest)
	}

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func CreateAppVersion(req *restful.Request, resp *restful.Response) {
	var createAppVersionRequest types.CreateAppVersionRequest
	err := req.ReadEntity(&createAppVersionRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	// override app id
	createAppVersionRequest.AppId = req.PathParameter("app")
	createAppVersionRequest.Username = req.HeaderParameter(constants.UserNameHeader)

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validatePackageRequest := &types.ValidatePackageRequest{
			VersionPackage: createAppVersionRequest.Package,
			VersionType:    createAppVersionRequest.Type,
		}
		result, err = app.ValidatePackage(validatePackageRequest)
	} else {
		result, err = app.CreateAppVersion(&createAppVersionRequest)
	}

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func ModifyAppVersion(req *restful.Request, resp *restful.Response) {

	var patchAppVersionRequest types.ModifyAppVersionRequest
	err := req.ReadEntity(&patchAppVersionRequest)
	versionId := req.PathParameter("version")

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = app.PatchAppVersion(versionId, &patchAppVersionRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DeleteAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")

	err := app.DeleteAppVersion(versionId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DescribeAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")

	result, err := app.DescribeAppVersion(versionId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}
