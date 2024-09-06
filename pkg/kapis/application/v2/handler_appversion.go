/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/net/context"
	"helm.sh/helm/v3/pkg/chart/loader"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/application"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

func (h *appHandler) CreateOrUpdateAppVersion(req *restful.Request, resp *restful.Response) {
	createAppVersionRequest, err := parseUpload(req)
	if requestDone(err, resp) {
		return
	}
	if h.ossStore == nil && len(createAppVersionRequest.Package) > maxFileSize {
		api.HandleBadRequest(resp, nil, fmt.Errorf("System has no OSS store, the maximum file size is %d", maxFileSize))
		return
	}

	createAppVersionRequest.AppName = req.PathParameter("app")

	vRequest, err := parseRequest(createAppVersionRequest)
	if requestDone(err, resp) {
		return
	}
	data := map[string]any{
		"icon":        vRequest.Icon,
		"appName":     vRequest.AppName,
		"versionName": vRequest.VersionName,
		"appHome":     vRequest.AppHome,
		"description": vRequest.Description,
		"aliasName":   vRequest.AliasName,
	}

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))
	if validate {
		resp.WriteAsJson(data)
		return
	}
	appVersion := &appv2.ApplicationVersion{}
	vRequest.VersionName = application.FormatVersion(vRequest.VersionName)
	appVersion.Name = fmt.Sprintf("%s-%s", createAppVersionRequest.AppName, vRequest.VersionName)
	if h.conflictedDone(req, resp, "version", appVersion) {
		return
	}

	app := appv2.Application{}
	err = h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: createAppVersionRequest.AppName}, &app)
	if requestDone(err, resp) {
		return
	}

	err = application.CreateOrUpdateAppVersion(req.Request.Context(), h.client, app, vRequest, h.cmStore, h.ossStore)
	if requestDone(err, resp) {
		return
	}
	err = application.UpdateLatestAppVersion(req.Request.Context(), h.client, app)
	if err != nil {
		klog.Errorf("failed to update latest app version, err:%v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(data)
}

func (h *appHandler) DeleteAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	err := h.client.Delete(req.Request.Context(), &appv2.ApplicationVersion{ObjectMeta: metav1.ObjectMeta{Name: versionId}})
	if requestDone(err, resp) {
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *appHandler) DescribeAppVersion(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	result := &appv2.ApplicationVersion{}
	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: versionId}, result)
	if requestDone(err, resp) {
		return
	}
	result.SetManagedFields(nil)

	resp.WriteEntity(result)
}

func (h *appHandler) ListAppVersions(req *restful.Request, resp *restful.Response) {

	var err error
	workspace := req.PathParameter("workspace")
	opt := runtimeclient.ListOptions{}
	labelSelectorStr := req.QueryParameter("labelSelector")
	labelSelector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	vals := []string{req.PathParameter("app")}
	appRequirement, err := labels.NewRequirement(appv2.AppIDLabelKey, selection.Equals, vals)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	labelSelector = labelSelector.Add(*appRequirement)
	opt.LabelSelector = labelSelector

	result := appv2.ApplicationVersionList{}
	err = h.client.List(req.Request.Context(), &result, &opt)

	if requestDone(err, resp) {
		return
	}
	conditions, err := params.ParseConditions(req)
	if requestDone(err, resp) {
		return
	}

	filtered := appv2.ApplicationVersionList{}
	for _, appv := range result.Items {
		states := strings.Split(conditions.Match[Status], "|")
		if conditions.Match[Status] != "" && !sliceutil.HasString(states, appv.Status.State) {
			continue
		}
		allowList := []string{appv2.SystemWorkspace, workspace}
		if workspace != "" && !stringutils.StringIn(appv.Labels[constants.WorkspaceLabelKey], allowList) {
			continue
		}

		filtered.Items = append(filtered.Items, appv)
	}

	resp.WriteEntity(convertToListResult(&filtered, req))
}

func (h *appHandler) GetAppVersionPackage(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")
	app := req.PathParameter("app")

	data, err := application.FailOverGet(h.cmStore, h.ossStore, versionId, h.client, true)
	if err != nil {
		klog.Errorf("get app version %s failed, error: %s", versionId, err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	result := map[string]any{
		"versionID": versionId,
		"package":   data,
		"appID":     app,
	}

	resp.WriteAsJson(result)
}

func (h *appHandler) GetAppVersionFiles(req *restful.Request, resp *restful.Response) {
	versionId := req.PathParameter("version")

	data, err := application.FailOverGet(h.cmStore, h.ossStore, versionId, h.client, true)
	if err != nil {
		klog.Errorf("get app version %s failed, error: %s", versionId, err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	var result = make(map[string][]byte)

	chartData, err := loader.LoadArchive(bytes.NewReader(data))
	if err != nil {
		result["all.yaml"] = data
		resp.WriteAsJson(result)
		return
	}

	for _, f := range chartData.Raw {
		result[f.Name] = f.Data
	}

	resp.WriteAsJson(result)
}

func (h *appHandler) AppVersionAction(req *restful.Request, resp *restful.Response) {
	versionID := req.PathParameter("version")
	var doActionRequest appv2.ApplicationVersionStatus
	err := req.ReadEntity(&doActionRequest)
	if requestDone(err, resp) {
		return
	}
	ctx := req.Request.Context()

	appVersion := &appv2.ApplicationVersion{}
	err = h.client.Get(ctx, runtimeclient.ObjectKey{Name: versionID}, appVersion)
	if requestDone(err, resp) {
		return
	}

	if doActionRequest.Message == "" {
		doActionRequest.Message = appVersion.Status.Message
	}

	// app version check state draft -> submitted -> (rejected -> submitted ) -> passed-> active -> (suspended -> active), draft -> submitted -> active

	err = DoAppVersionAction(ctx, versionID, doActionRequest, h.client)
	if requestDone(err, resp) {
		return
	}

	resp.WriteEntity(errors.None)
}

func DoAppVersionAction(ctx context.Context, versionId string, actionReq appv2.ApplicationVersionStatus, client runtimeclient.Client) error {

	key := runtimeclient.ObjectKey{Name: versionId}
	version := &appv2.ApplicationVersion{}
	err := client.Get(ctx, key, version)
	if err != nil {
		klog.Errorf("get app version %s failed, error: %s", versionId, err)
		return err
	}
	version.Status.State = actionReq.State
	if actionReq.Message != "" {
		version.Status.Message = actionReq.Message
	}

	if actionReq.UserName != "" {
		version.Status.UserName = actionReq.UserName
	}
	version.Status.Updated = &metav1.Time{Time: metav1.Now().Time}
	err = client.Status().Update(ctx, version)
	if err != nil {
		klog.Errorf("update app version %s failed, error: %s", versionId, err)
	}

	appID := version.Labels[appv2.AppIDLabelKey]
	err = checkAppStatus(ctx, appID, actionReq.State, client)
	if err != nil {
		klog.Errorf("check app failed, error: %s", err)
		return err
	}

	return err
}

func checkAppStatus(ctx context.Context, appID, action string, client runtimeclient.Client) error {
	//If all appVersions are Suspended, then the app's status is not active
	if action != appv2.ReviewStatusSuspended {
		return nil
	}

	allVersion := appv2.ApplicationVersionList{}
	opt := runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.AppIDLabelKey: appID}),
	}
	err := client.List(ctx, &allVersion, &opt)
	if err != nil {
		klog.Errorf("list app versions failed, error: %s", err)
		return err
	}
	active := false
	for _, v := range allVersion.Items {
		if v.Status.State == appv2.ReviewStatusActive {
			active = true
			break
		}
	}
	app := appv2.Application{}
	err = client.Get(ctx, runtimeclient.ObjectKey{Name: appID}, &app)
	if err != nil {
		klog.Errorf("get app %s failed, error: %s", appID, err)
		return err
	}

	if !active && app.Status.State == appv2.ReviewStatusActive {
		app.Status.State = appv2.ReviewStatusDraft
		err = client.Status().Update(ctx, &app)
		if err != nil {
			klog.Errorf("update app %s failed, error: %s", appID, err)
			return err
		}
	}
	return err
}

func (h *appHandler) ListReviews(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req)

	if requestDone(err, resp) {
		return
	}

	appVersions := appv2.ApplicationVersionList{}
	opts := runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.RepoIDLabelKey: appv2.UploadRepoKey}),
	}
	err = h.client.List(req.Request.Context(), &appVersions, &opts)
	if requestDone(err, resp) {
		return
	}

	if conditions == nil || len(conditions.Match) == 0 {
		resp.WriteEntity(convertToListResult(&appVersions, req))
		return
	}

	filteredAppVersions := appv2.ApplicationVersionList{}
	states := strings.Split(conditions.Match[Status], "|")
	for _, version := range appVersions.Items {
		if conditions.Match[Status] != "" && !sliceutil.HasString(states, version.Status.State) {
			continue
		}
		filteredAppVersions.Items = append(filteredAppVersions.Items, version)
	}

	resp.WriteEntity(convertToListResult(&filteredAppVersions, req))
}
