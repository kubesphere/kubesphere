/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"encoding/json"
	"fmt"
	"io"

	"mime"
	"strconv"
	"strings"
	"time"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"

	"kubesphere.io/kubesphere/pkg/utils/sliceutil"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/constants"

	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/application"
)

const maxFileSize = 1 * 1024 * 1024 // 1 MB in bytes

func (h *appHandler) CreateOrUpdateApp(req *restful.Request, resp *restful.Response) {
	createAppRequest, err := parseUpload(req)
	if requestDone(err, resp) {
		return
	}

	if h.ossStore == nil && len(createAppRequest.Package) > maxFileSize {
		api.HandleBadRequest(resp, nil, fmt.Errorf("System has no OSS store, the maximum file size is %d", maxFileSize))
		return
	}

	newReq, err := parseRequest(createAppRequest)
	if requestDone(err, resp) {
		return
	}
	data := map[string]any{
		"icon":        newReq.Icon,
		"appName":     newReq.AppName,
		"versionName": newReq.VersionName,
		"appHome":     newReq.AppHome,
		"description": newReq.Description,
		"aliasName":   newReq.AliasName,
		"resources":   newReq.Resources,
	}

	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))
	if validate {
		resp.WriteAsJson(data)
		return
	}

	app := &appv2.Application{}
	app.Name = newReq.AppName
	if h.conflictedDone(req, resp, "app", app) {
		return
	}

	newReq.FromRepo = false
	vRequests := []application.AppRequest{newReq}
	err = application.CreateOrUpdateApp(h.client, vRequests, h.cmStore, h.ossStore)
	if requestDone(err, resp) {
		return
	}

	resp.WriteAsJson(data)
}

func parseUpload(req *restful.Request) (createAppRequest application.AppRequest, err error) {
	contentType := req.Request.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		klog.Errorf("parse media type failed, err: %s", err)
		return createAppRequest, err
	}
	if mediaType == "multipart/form-data" {
		file, header, err := req.Request.FormFile("file")
		if err != nil {
			klog.Errorf("parse form file failed, err: %s", err)
			return createAppRequest, err
		}
		klog.Info("upload file:", header.Filename)
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			klog.Errorf("read file failed, err: %s", err)
			return createAppRequest, err
		}
		err = json.Unmarshal([]byte(req.Request.FormValue("jsonData")), &createAppRequest)
		if err != nil {
			klog.Errorf("parse json data failed, err: %s", err)
			return createAppRequest, err
		}
		if createAppRequest.Package != nil {
			return createAppRequest, errors.New("When using multipart/form-data to upload files, the package field does not need to be included in the json data.")
		}
		createAppRequest.Package = data
		return createAppRequest, nil
	}
	if mediaType == "application/json" {
		err = req.ReadEntity(&createAppRequest)
		if err != nil {
			klog.Errorf("parse json data failed, err: %s", err)
			return createAppRequest, err
		}
		return createAppRequest, nil
	}
	return createAppRequest, errors.New("unsupported media type")
}

func (h *appHandler) ListApps(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	conditions, err := params.ParseConditions(req)
	if requestDone(err, resp) {
		return
	}

	opt := runtimeclient.ListOptions{}
	labelSelectorStr := req.QueryParameter("labelSelector")
	labelSelector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	opt.LabelSelector = labelSelector

	result := appv2.ApplicationList{}
	err = h.client.List(req.Request.Context(), &result, &opt)
	if requestDone(err, resp) {
		return
	}

	filtered := appv2.ApplicationList{}
	for _, app := range result.Items {
		curApp := app
		states := strings.Split(conditions.Match[Status], "|")
		if conditions.Match[Status] != "" && !sliceutil.HasString(states, curApp.Status.State) {
			continue
		}
		allowList := []string{appv2.SystemWorkspace, workspace}
		if workspace != "" && !stringutils.StringIn(app.Labels[constants.WorkspaceLabelKey], allowList) {
			continue
		}
		filtered.Items = append(filtered.Items, curApp)
	}

	resp.WriteEntity(convertToListResult(&filtered, req))
}

func (h *appHandler) DescribeApp(req *restful.Request, resp *restful.Response) {

	key := runtimeclient.ObjectKey{Name: req.PathParameter("app")}
	app := &appv2.Application{}
	err := h.client.Get(req.Request.Context(), key, app)
	if requestDone(err, resp) {
		return
	}
	app.SetManagedFields(nil)

	resp.WriteEntity(app)
}

func (h *appHandler) DeleteApp(req *restful.Request, resp *restful.Response) {
	appId := req.PathParameter("app")
	app := &appv2.Application{}

	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: appId}, app)
	if requestDone(err, resp) {
		return
	}

	err = application.FailOverDelete(h.cmStore, h.ossStore, app.Spec.Attachments)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	err = h.client.Delete(req.Request.Context(), &appv2.Application{ObjectMeta: metav1.ObjectMeta{Name: appId}})
	if requestDone(err, resp) {
		return
	}

	resp.WriteEntity(errors.None)
}

func (h *appHandler) DoAppAction(req *restful.Request, resp *restful.Response) {
	var doActionRequest appv2.ApplicationVersionStatus
	err := req.ReadEntity(&doActionRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	ctx := req.Request.Context()

	user, _ := request.UserFrom(req.Request.Context())
	if user != nil {
		doActionRequest.UserName = user.GetName()
	}
	app := &appv2.Application{}
	app.Name = req.PathParameter("app")

	err = h.client.Get(ctx, runtimeclient.ObjectKey{Name: app.Name}, app)
	if requestDone(err, resp) {
		return
	}

	// app state check, draft -> active -> suspended -> active
	switch doActionRequest.State {
	case appv2.ReviewStatusActive:
		if app.Status.State != appv2.ReviewStatusDraft &&
			app.Status.State != appv2.ReviewStatusSuspended {
			err = fmt.Errorf("app %s is not in draft or suspended status", app.Name)
			break
		}
		// active state is only allowed if at least one app version is in active or passed state
		appVersionList := &appv2.ApplicationVersionList{}
		opt := &runtimeclient.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{appv2.AppIDLabelKey: app.Name})}
		h.client.List(ctx, appVersionList, opt)
		okActive := false
		if len(appVersionList.Items) > 0 {
			for _, v := range appVersionList.Items {
				if v.Status.State == appv2.ReviewStatusActive ||
					v.Status.State == appv2.ReviewStatusPassed {
					okActive = true
					break
				}
			}
		}
		if !okActive {
			err = fmt.Errorf("app %s has no active or passed appversion", app.Name)
		}
	case appv2.ReviewStatusSuspended:
		if app.Status.State != appv2.ReviewStatusActive {
			err = fmt.Errorf("app %s is not in active status", app.Name)
		}
	}

	if requestDone(err, resp) {
		return
	}
	if doActionRequest.State == appv2.ReviewStatusActive {
		if app.Labels == nil {
			app.Labels = map[string]string{}
		}
		app.Labels[appv2.AppStoreLabelKey] = "true"
		h.client.Update(ctx, app)
	}

	app.Status.State = doActionRequest.State
	app.Status.UpdateTime = &metav1.Time{Time: time.Now()}
	err = h.client.Status().Update(ctx, app)
	if requestDone(err, resp) {
		return
	}

	//update appversion status
	versions := &appv2.ApplicationVersionList{}

	opt := &runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{appv2.AppIDLabelKey: app.Name}),
	}

	err = h.client.List(ctx, versions, opt)
	if requestDone(err, resp) {
		return
	}
	for _, version := range versions.Items {
		if version.Status.State == appv2.StatusActive || version.Status.State == appv2.ReviewStatusSuspended {
			err = DoAppVersionAction(ctx, version.Name, doActionRequest, h.client)
			if err != nil {
				klog.V(4).Infoln(err)
				api.HandleInternalError(resp, nil, err)
				return
			}
		}
	}

	resp.WriteEntity(errors.None)
}

func (h *appHandler) PatchApp(req *restful.Request, resp *restful.Response) {
	var err error
	appId := req.PathParameter("app")
	appBody := &application.AppRequest{}
	err = req.ReadEntity(appBody)
	if requestDone(err, resp) {
		return
	}
	ctx := req.Request.Context()

	app := &appv2.Application{}
	app.Name = appId
	err = h.client.Get(ctx, runtimeclient.ObjectKey{Name: app.Name}, app)
	if requestDone(err, resp) {
		return
	}
	if app.GetLabels() == nil {
		app.SetLabels(map[string]string{})
	}
	app.Labels[appv2.AppCategoryNameKey] = appBody.CategoryName
	app.Spec.Icon = appBody.Icon
	ant := app.GetAnnotations()
	if ant == nil {
		ant = make(map[string]string)
	}

	ant[constants.DescriptionAnnotationKey] = appBody.Description
	ant[constants.DisplayNameAnnotationKey] = appBody.AliasName
	app.SetAnnotations(ant)

	app.Spec.Attachments = appBody.Attachments
	app.Spec.Abstraction = appBody.Abstraction
	app.Spec.AppHome = appBody.AppHome

	err = h.client.Update(ctx, app)
	if requestDone(err, resp) {
		return
	}
	klog.V(4).Infof("update app %s successfully", app.Name)

	resp.WriteAsJson(app)
}
