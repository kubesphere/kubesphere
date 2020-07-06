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

package openpitrix

import (
	"github.com/go-openapi/strfmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"strings"
)

type AppTemplateInterface interface {
	ListApps(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	DescribeApp(id string) (*App, error)
	DeleteApp(id string) error
	CreateApp(request *CreateAppRequest) (*CreateAppResponse, error)
	ModifyApp(appId string, request *ModifyAppRequest) error
	DeleteAppVersion(id string) error
	ModifyAppVersion(id string, request *ModifyAppVersionRequest) error
	DescribeAppVersion(id string) (*AppVersion, error)
	CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error)
	ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error)
	GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error)
	DoAppAction(appId string, request *ActionRequest) error
	DoAppVersionAction(versionId string, request *ActionRequest) error
	GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error)
	ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
}

type appTemplateOperator struct {
	opClient openpitrix.Client
}

func newAppTemplateOperator(opClient openpitrix.Client) AppTemplateInterface {
	return &appTemplateOperator{
		opClient: opClient,
	}
}

func (c *appTemplateOperator) ListApps(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	describeAppsRequest := &pb.DescribeAppsRequest{}
	if keyword := conditions.Match[Keyword]; keyword != "" {
		describeAppsRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match[AppId]; appId != "" {
		describeAppsRequest.AppId = strings.Split(appId, "|")
	}
	if isv := conditions.Match[ISV]; isv != "" {
		describeAppsRequest.Isv = strings.Split(isv, "|")
	}
	if categoryId := conditions.Match[CategoryId]; categoryId != "" {
		describeAppsRequest.CategoryId = strings.Split(categoryId, "|")
	}
	if repoId := conditions.Match[RepoId]; repoId != "" {
		// hard code, app template in built-in repo has no repo_id attribute
		if repoId == BuiltinRepoId {
			describeAppsRequest.RepoId = []string{"\u0000"}
		} else {
			describeAppsRequest.RepoId = strings.Split(repoId, "|")
		}
	}
	if status := conditions.Match[Status]; status != "" {
		describeAppsRequest.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppsRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppsRequest.Reverse = &wrappers.BoolValue{Value: reverse}
	describeAppsRequest.Limit = uint32(limit)
	describeAppsRequest.Offset = uint32(offset)
	resp, err := c.opClient.DescribeApps(openpitrix.SystemContext(), describeAppsRequest)
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

func (c *appTemplateOperator) DescribeApp(id string) (*App, error) {
	resp, err := c.opClient.DescribeApps(openpitrix.SystemContext(), &pb.DescribeAppsRequest{
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

func (c *appTemplateOperator) DeleteApp(id string) error {
	_, err := c.opClient.DeleteApps(openpitrix.SystemContext(), &pb.DeleteAppsRequest{
		AppId: []string{id},
	})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *appTemplateOperator) CreateApp(request *CreateAppRequest) (*CreateAppResponse, error) {
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
	resp, err := c.opClient.CreateApp(openpitrix.ContextWithUsername(request.Username), createAppRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &CreateAppResponse{
		AppID:     resp.GetAppId().GetValue(),
		VersionID: resp.GetVersionId().GetValue(),
	}, nil
}

func (c *appTemplateOperator) ModifyApp(appId string, request *ModifyAppRequest) error {
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

		_, err := c.opClient.UploadAppAttachment(openpitrix.SystemContext(), uploadAttachmentRequest)

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

		_, err := c.opClient.ModifyApp(openpitrix.SystemContext(), patchAppRequest)

		if err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func (c *appTemplateOperator) CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error) {
	createAppVersionRequest := &pb.CreateAppVersionRequest{
		AppId:       &wrappers.StringValue{Value: request.AppId},
		Name:        &wrappers.StringValue{Value: request.Name},
		Description: &wrappers.StringValue{Value: request.Description},
		Type:        &wrappers.StringValue{Value: request.Type},
	}

	if request.Package != nil {
		createAppVersionRequest.Package = &wrappers.BytesValue{Value: request.Package}
	}

	resp, err := c.opClient.CreateAppVersion(openpitrix.ContextWithUsername(request.Username), createAppVersionRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &CreateAppVersionResponse{
		VersionId: resp.GetVersionId().GetValue(),
	}, nil
}

func (c *appTemplateOperator) ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error) {
	r := &pb.ValidatePackageRequest{}

	if request.VersionPackage != nil {
		r.VersionPackage = request.VersionPackage
	}
	if request.VersionType != "" {
		r.VersionType = request.VersionType
	}

	resp, err := c.opClient.ValidatePackage(openpitrix.SystemContext(), r)

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

func (c *appTemplateOperator) DeleteAppVersion(id string) error {
	_, err := c.opClient.DeleteAppVersion(openpitrix.SystemContext(), &pb.DeleteAppVersionRequest{
		VersionId: &wrappers.StringValue{Value: id},
	})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *appTemplateOperator) ModifyAppVersion(id string, request *ModifyAppVersionRequest) error {
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

	_, err := c.opClient.ModifyAppVersion(openpitrix.SystemContext(), modifyAppVersionRequest)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *appTemplateOperator) DescribeAppVersion(id string) (*AppVersion, error) {
	resp, err := c.opClient.DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
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

func (c *appTemplateOperator) GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error) {
	resp, err := c.opClient.GetAppVersionPackage(openpitrix.SystemContext(), &pb.GetAppVersionPackageRequest{
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

func (c *appTemplateOperator) DoAppAction(appId string, request *ActionRequest) error {
	switch request.Action {

	case ActionRecover:
		// TODO openpitrix need to implement app recover interface
		resp, err := c.opClient.DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
			AppId:  []string{appId},
			Status: []string{StatusSuspended},
			Limit:  200,
			Offset: 0,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		for _, version := range resp.AppVersionSet {

			_, err = c.opClient.RecoverAppVersion(openpitrix.SystemContext(), &pb.RecoverAppVersionRequest{
				VersionId: version.VersionId,
			})
			if err != nil {
				klog.Error(err)
				return err
			}
		}

	case ActionSuspend:
		// TODO openpitrix need to implement app suspend interface
		resp, err := c.opClient.DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{
			AppId:  []string{appId},
			Status: []string{StatusActive},
			Limit:  200,
			Offset: 0,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		for _, version := range resp.AppVersionSet {
			_, err = c.opClient.SuspendAppVersion(openpitrix.SystemContext(), &pb.SuspendAppVersionRequest{
				VersionId: version.VersionId,
			})

			if err != nil {
				klog.Error(err)
				return err
			}
		}

	default:
		err := status.New(codes.InvalidArgument, "action not support").Err()
		klog.Error(err)
		return err
	}

	return nil
}

func (c *appTemplateOperator) DoAppVersionAction(versionId string, request *ActionRequest) error {
	var err error
	switch request.Action {
	case ActionCancel:
		_, err = c.opClient.CancelAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.CancelAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case ActionPass:
		_, err = c.opClient.AdminPassAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.PassAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case ActionRecover:
		_, err = c.opClient.RecoverAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.RecoverAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case ActionReject:
		_, err = c.opClient.AdminRejectAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.RejectAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
			Message:   &wrappers.StringValue{Value: request.Message},
		})
	case ActionSubmit:
		_, err = c.opClient.SubmitAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.SubmitAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case ActionSuspend:
		_, err = c.opClient.SuspendAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.SuspendAppVersionRequest{
			VersionId: &wrappers.StringValue{Value: versionId},
		})
	case ActionRelease:
		_, err = c.opClient.ReleaseAppVersion(openpitrix.ContextWithUsername(request.Username), &pb.ReleaseAppVersionRequest{
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

func (c *appTemplateOperator) GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error) {
	getAppVersionPackageFilesRequest := &pb.GetAppVersionPackageFilesRequest{
		VersionId: &wrappers.StringValue{Value: versionId},
	}
	if request.Files != nil {
		getAppVersionPackageFilesRequest.Files = request.Files
	}

	resp, err := c.opClient.GetAppVersionPackageFiles(openpitrix.SystemContext(), getAppVersionPackageFilesRequest)
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

func (c *appTemplateOperator) ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	describeAppVersionAudits := &pb.DescribeAppVersionAuditsRequest{}

	if keyword := conditions.Match[Keyword]; keyword != "" {
		describeAppVersionAudits.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match[AppId]; appId != "" {
		describeAppVersionAudits.AppId = []string{appId}
	}
	if versionId := conditions.Match[VersionId]; versionId != "" {
		describeAppVersionAudits.VersionId = []string{versionId}
	}
	if status := conditions.Match[Status]; status != "" {
		describeAppVersionAudits.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionAudits.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionAudits.Reverse = &wrappers.BoolValue{Value: reverse}
	describeAppVersionAudits.Limit = uint32(limit)
	describeAppVersionAudits.Offset = uint32(offset)
	resp, err := c.opClient.DescribeAppVersionAudits(openpitrix.SystemContext(), describeAppVersionAudits)

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

func (c *appTemplateOperator) ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	describeAppVersionReviews := &pb.DescribeAppVersionReviewsRequest{}

	if keyword := conditions.Match[Keyword]; keyword != "" {
		describeAppVersionReviews.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if status := conditions.Match[Status]; status != "" {
		describeAppVersionReviews.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionReviews.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionReviews.Reverse = &wrappers.BoolValue{Value: reverse}
	describeAppVersionReviews.Limit = uint32(limit)
	describeAppVersionReviews.Offset = uint32(offset)
	// TODO icon is needed
	resp, err := c.opClient.DescribeAppVersionReviews(openpitrix.SystemContext(), describeAppVersionReviews)

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

func (c *appTemplateOperator) ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	describeAppVersionsRequest := &pb.DescribeAppVersionsRequest{}

	if keyword := conditions.Match[Keyword]; keyword != "" {
		describeAppVersionsRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if appId := conditions.Match[AppId]; appId != "" {
		describeAppVersionsRequest.AppId = []string{appId}
	}
	if status := conditions.Match[Status]; status != "" {
		describeAppVersionsRequest.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeAppVersionsRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeAppVersionsRequest.Reverse = &wrappers.BoolValue{Value: reverse}
	describeAppVersionsRequest.Limit = uint32(limit)
	describeAppVersionsRequest.Offset = uint32(offset)
	resp, err := c.opClient.DescribeAppVersions(openpitrix.SystemContext(), describeAppVersionsRequest)

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
