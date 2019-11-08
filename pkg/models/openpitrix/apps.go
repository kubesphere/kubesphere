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
	"github.com/go-openapi/strfmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"strings"
)

const (
	BuiltinRepoId = "repo-helm"
	StatusActive  = "active"
)

func ListApps(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	describeAppsRequest := &pb.DescribeAppsRequest{}
	if keyword := conditions.Match["keyword"]; keyword != "" {
		describeAppsRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match["app_id"]; appId != "" {
		describeAppsRequest.AppId = strings.Split(appId, "|")
	}
	if isv := conditions.Match["isv"]; isv != "" {
		describeAppsRequest.Isv = strings.Split(isv, "|")
	}
	if categoryId := conditions.Match["category_id"]; categoryId != "" {
		describeAppsRequest.CategoryId = strings.Split(categoryId, "|")
	}
	if repoId := conditions.Match["repo"]; repoId != "" {
		// hard code, app template in built-in repo has no repo_id attribute
		if repoId == BuiltinRepoId {
			describeAppsRequest.RepoId = []string{"\u0000"}
		} else {
			describeAppsRequest.RepoId = strings.Split(repoId, "|")
		}
	}
	if status := conditions.Match["status"]; status != "" {
		describeAppsRequest.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppsRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppsRequest.Reverse = &wrappers.BoolValue{Value: reverse}
	describeAppsRequest.Limit = uint32(limit)
	describeAppsRequest.Offset = uint32(offset)
	resp, err := client.App().DescribeApps(openpitrix.SystemContext(), describeAppsRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.AppSet {
		items = append(items, convertApp(item))
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}

func DescribeApp(id string) (*App, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	resp, err := op.App().DescribeApps(openpitrix.SystemContext(), &pb.DescribeAppsRequest{
		AppId: []string{id},
		Limit: 1,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var app *App

	if len(resp.AppSet) > 0 {
		app = convertApp(resp.AppSet[0])
		return app, nil
	} else {
		err := status.New(codes.NotFound, "resource not found").Err()
		klog.Error(err)
		return nil, err
	}
}

func DeleteApp(id string) error {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return err
	}
	_, err = op.App().DeleteApps(openpitrix.SystemContext(), &pb.DeleteAppsRequest{
		AppId: []string{id},
	})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func CreateApp(request *CreateAppRequest) (*CreateAppResponse, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	createAppRequest := &pb.CreateAppRequest{
		Name:        &wrappers.StringValue{Value: request.Name},
		VersionType: &wrappers.StringValue{Value: request.VersionType},
		VersionName: &wrappers.StringValue{Value: request.VersionName},
	}
	if request.VersionPackage != nil {
		createAppRequest.VersionPackage = &wrappers.BytesValue{Value: request.VersionPackage}
	}
	if request.Icon != nil {
		createAppRequest.Icon = &wrappers.BytesValue{Value: request.Icon}
	}
	if request.Isv != "" {
		createAppRequest.Isv = &wrappers.StringValue{Value: request.Isv}
	}
	resp, err := op.App().CreateApp(openpitrix.ContextWithUsername(request.Username), createAppRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &CreateAppResponse{
		AppID:     resp.GetAppId().GetValue(),
		VersionID: resp.GetVersionId().GetValue(),
	}, nil
}

func PatchApp(appId string, request *ModifyAppRequest) error {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	// upload app attachment
	if request.AttachmentContent != nil {
		uploadAttachmentRequest := &pb.UploadAppAttachmentRequest{
			AppId:             &wrappers.StringValue{Value: appId},
			AttachmentContent: &wrappers.BytesValue{Value: request.AttachmentContent},
		}
		if request.Type != nil {
			uploadAttachmentRequest.Type = pb.UploadAppAttachmentRequest_Type(pb.UploadAppAttachmentRequest_Type_value[*request.Type])
		}
		if request.Sequence != nil {
			uploadAttachmentRequest.Sequence = &wrappers.UInt32Value{Value: uint32(*request.Sequence)}
		}

		_, err := client.App().UploadAppAttachment(openpitrix.SystemContext(), uploadAttachmentRequest)

		if err != nil {
			klog.Error(err)
			return err
		}
		// patch app
	} else {
		patchAppRequest := &pb.ModifyAppRequest{
			AppId: &wrappers.StringValue{Value: appId},
		}

		if request.Abstraction != nil {
			patchAppRequest.Abstraction = &wrappers.StringValue{Value: *request.Abstraction}
		}
		if request.CategoryID != nil {
			patchAppRequest.CategoryId = &wrappers.StringValue{Value: *request.CategoryID}
		}
		if request.Description != nil {
			patchAppRequest.Description = &wrappers.StringValue{Value: *request.Description}
		}
		if request.Home != nil {
			patchAppRequest.Home = &wrappers.StringValue{Value: *request.Home}
		}
		if request.Keywords != nil {
			patchAppRequest.Keywords = &wrappers.StringValue{Value: *request.Keywords}
		}
		if request.Maintainers != nil {
			patchAppRequest.Maintainers = &wrappers.StringValue{Value: *request.Maintainers}
		}
		if request.Name != nil {
			patchAppRequest.Name = &wrappers.StringValue{Value: *request.Name}
		}
		if request.Readme != nil {
			patchAppRequest.Readme = &wrappers.StringValue{Value: *request.Readme}
		}
		if request.Sources != nil {
			patchAppRequest.Sources = &wrappers.StringValue{Value: *request.Sources}
		}
		if request.Tos != nil {
			patchAppRequest.Tos = &wrappers.StringValue{Value: *request.Tos}
		}

		_, err = client.App().ModifyApp(openpitrix.SystemContext(), patchAppRequest)

		if err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	createAppVersionRequest := &pb.CreateAppVersionRequest{
		AppId:       &wrappers.StringValue{Value: request.AppId},
		Name:        &wrappers.StringValue{Value: request.Name},
		Description: &wrappers.StringValue{Value: request.Description},
		Type:        &wrappers.StringValue{Value: request.Type},
	}

	if request.Package != nil {
		createAppVersionRequest.Package = &wrappers.BytesValue{Value: request.Package}
	}

	resp, err := op.App().CreateAppVersion(openpitrix.ContextWithUsername(request.Username), createAppVersionRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &CreateAppVersionResponse{
		VersionId: resp.GetVersionId().GetValue(),
	}, nil
}

func ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	r := &pb.ValidatePackageRequest{}

	if request.VersionPackage != nil {
		r.VersionPackage = request.VersionPackage
	}
	if request.VersionType != "" {
		r.VersionType = request.VersionType
	}

	resp, err := client.App().ValidatePackage(openpitrix.SystemContext(), r)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := &ValidatePackageResponse{}

	if resp.Error != nil {
		result.Error = resp.Error.Value
	}
	if resp.Description != nil {
		result.Description = resp.Description.Value
	}
	if resp.Error != nil {
		result.Error = resp.Error.Value
	}
	if resp.ErrorDetails != nil {
		result.ErrorDetails = resp.ErrorDetails
	}
	if resp.Name != nil {
		result.Name = resp.Name.Value
	}
	if resp.Url != nil {
		result.URL = resp.Url.Value
	}
	if resp.VersionName != nil {
		result.VersionName = resp.VersionName.Value
	}
	return result, nil
}

func DeleteAppVersion(id string) error {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return err
	}
	_, err = op.App().DeleteAppVersion(openpitrix.SystemContext(), &pb.DeleteAppVersionRequest{
		VersionId: &wrappers.StringValue{Value: id},
	})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func PatchAppVersion(id string, request *ModifyAppVersionRequest) error {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return err
	}
	modifyAppVersionRequest := &pb.ModifyAppVersionRequest{
		VersionId: &wrappers.StringValue{Value: id},
	}

	if request.Name != nil {
		modifyAppVersionRequest.Name = &wrappers.StringValue{Value: *request.Name}
	}
	if request.Description != nil {
		modifyAppVersionRequest.Description = &wrappers.StringValue{Value: *request.Description}
	}
	if request.Package != nil {
		modifyAppVersionRequest.Package = &wrappers.BytesValue{Value: request.Package}
	}
	if request.PackageFiles != nil {
		modifyAppVersionRequest.PackageFiles = request.PackageFiles
	}

	_, err = op.App().ModifyAppVersion(openpitrix.SystemContext(), modifyAppVersionRequest)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func DescribeAppVersion(id string) (*AppVersion, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	resp, err := op.App().DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
		VersionId: []string{id},
		Limit:     1,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var app *AppVersion

	if len(resp.AppVersionSet) > 0 {
		app = convertAppVersion(resp.AppVersionSet[0])
		return app, nil
	} else {
		err := status.New(codes.NotFound, "resource not found").Err()
		klog.Error(err)
		return nil, err
	}
}

func GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	resp, err := op.App().GetAppVersionPackage(openpitrix.SystemContext(), &pb.GetAppVersionPackageRequest{
		VersionId: &wrappers.StringValue{Value: versionId},
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	app := &GetAppVersionPackageResponse{
		AppId:     appId,
		VersionId: versionId,
	}

	if resp.Package != nil {
		app.Package = resp.Package
	}

	return app, nil
}

func DoAppAction(appId string, request *ActionRequest) error {
	op, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	switch request.Action {

	case "recover":
		// TODO openpitrix need to implement app recover interface
		resp, err := op.App().DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
			AppId:  []string{appId},
			Status: []string{"suspended"},
			Limit:  200,
			Offset: 0,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		for _, version := range resp.AppVersionSet {

			_, err = op.App().RecoverAppVersion(openpitrix.SystemContext(), &pb.RecoverAppVersionRequest{
				VersionId: version.VersionId,
			})
			if err != nil {
				klog.Error(err)
				return err
			}
		}

	case "suspend":
		// TODO openpitrix need to implement app suspend interface
		resp, err := op.App().DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
			AppId:  []string{appId},
			Status: []string{"active"},
			Limit:  200,
			Offset: 0,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		for _, version := range resp.AppVersionSet {
			_, err = op.App().SuspendAppVersion(openpitrix.SystemContext(), &pb.SuspendAppVersionRequest{
				VersionId: version.VersionId,
			})

			if err != nil {
				klog.Error(err)
				return err
			}
		}

	default:
		err = status.New(codes.InvalidArgument, "action not support").Err()
		klog.Error(err)
		return err
	}

	return nil
}

func DoAppVersionAction(versionId string, request *ActionRequest) error {
	op, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	switch request.Action {
	case "cancel":
		_, err = op.App().CancelAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.CancelAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case "pass":
		_, err = op.App().AdminPassAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.PassAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case "recover":
		_, err = op.App().RecoverAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.RecoverAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case "reject":
		_, err = op.App().AdminRejectAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.RejectAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
			Message:   &wrappers.StringValue{Value: request.Message},
		})
	case "submit":
		_, err = op.App().SubmitAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.SubmitAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case "suspend":
		_, err = op.App().SuspendAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.SuspendAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case "release":
		_, err = op.App().ReleaseAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.ReleaseAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	default:
		err = status.New(codes.InvalidArgument, "action not support").Err()
	}

	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	getAppVersionPackageFilesRequest := &pb.GetAppVersionPackageFilesRequest{
		VersionId: &wrappers.StringValue{Value: versionId},
	}
	if request.Files != nil {
		getAppVersionPackageFilesRequest.Files = request.Files
	}

	resp, err := op.App().GetAppVersionPackageFiles(openpitrix.SystemContext(), getAppVersionPackageFilesRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	version := &GetAppVersionPackageFilesResponse{
		VersionId: versionId,
	}

	if resp.Files != nil {
		version.Files = make(map[string]strfmt.Base64)
		for k, v := range resp.Files {
			version.Files[k] = v
		}
	}

	return version, nil
}

func ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	describeAppVersionAudits := &pb.DescribeAppVersionAuditsRequest{}

	if keyword := conditions.Match["keyword"]; keyword != "" {
		describeAppVersionAudits.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match["app"]; appId != "" {
		describeAppVersionAudits.AppId = []string{appId}
	}
	if versionId := conditions.Match["version"]; versionId != "" {
		describeAppVersionAudits.VersionId = []string{versionId}
	}
	if status := conditions.Match["status"]; status != "" {
		describeAppVersionAudits.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionAudits.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionAudits.Reverse = &wrappers.BoolValue{Value: !reverse}
	describeAppVersionAudits.Limit = uint32(limit)
	describeAppVersionAudits.Offset = uint32(offset)
	resp, err := client.App().DescribeAppVersionAudits(openpitrix.SystemContext(), describeAppVersionAudits)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.AppVersionAuditSet {
		appVersion := convertAppVersionAudit(item)
		items = append(items, appVersion)
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}

func ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	describeAppVersionReviews := &pb.DescribeAppVersionReviewsRequest{}

	if keyword := conditions.Match["keyword"]; keyword != "" {
		describeAppVersionReviews.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if status := conditions.Match["status"]; status != "" {
		describeAppVersionReviews.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionReviews.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionReviews.Reverse = &wrappers.BoolValue{Value: !reverse}
	describeAppVersionReviews.Limit = uint32(limit)
	describeAppVersionReviews.Offset = uint32(offset)
	// TODO icon is needed
	resp, err := client.App().DescribeAppVersionReviews(openpitrix.SystemContext(), describeAppVersionReviews)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.AppVersionReviewSet {
		appVersion := convertAppVersionReview(item)
		items = append(items, appVersion)
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}

func ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	describeAppVersionsRequest := &pb.DescribeAppVersionsRequest{}

	if keyword := conditions.Match["keyword"]; keyword != "" {
		describeAppVersionsRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match["app"]; appId != "" {
		describeAppVersionsRequest.AppId = []string{appId}
	}
	if status := conditions.Match["status"]; status != "" {
		describeAppVersionsRequest.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionsRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionsRequest.Reverse = &wrappers.BoolValue{Value: !reverse}
	describeAppVersionsRequest.Limit = uint32(limit)
	describeAppVersionsRequest.Offset = uint32(offset)
	resp, err := client.App().DescribeAppVersions(openpitrix.SystemContext(), describeAppVersionsRequest)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.AppVersionSet {
		appVersion := convertAppVersion(item)
		items = append(items, appVersion)
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}
