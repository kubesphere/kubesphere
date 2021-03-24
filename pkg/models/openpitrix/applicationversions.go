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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/strfmt"
	"io"
	"io/ioutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"math"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
)

func (c *applicationOperator) GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error) {
	var version *v1alpha1.HelmApplicationVersion
	var err error

	version, err = c.getAppVersionByVersionIdWithData(versionId)
	if err != nil {
		return nil, err
	}

	return &GetAppVersionPackageResponse{
		AppId:     appId,
		VersionId: versionId,
		Package:   version.Spec.Data,
	}, nil
}

// check helm package and create helm app version if not exist
func (c *applicationOperator) CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error) {
	if c.backingStoreClient == nil {
		return nil, invalidS3Config
	}

	chrt, err := helmrepoindex.LoadPackage(request.Package)
	if err != nil {
		klog.Errorf("load package failed, error: %s", err)
		return nil, err
	}

	app, err := c.appLister.Get(request.AppId)

	if err != nil {
		klog.Errorf("get app %s failed, error: %s", request.AppId, err)
		return nil, err
	}
	chartPackage := request.Package.String()
	version := buildApplicationVersion(app, chrt, &chartPackage, request.Username)
	version, err = c.createApplicationVersion(version)

	if err != nil {
		klog.Errorf("create helm app version failed, error: %s", err)
		return nil, err
	}

	klog.V(4).Infof("create helm app version %s success", request.Name)

	return &CreateAppVersionResponse{
		VersionId: version.GetHelmApplicationVersionId(),
	}, nil
}

func (c *applicationOperator) DeleteAppVersion(id string) error {
	appVersion, err := c.versionLister.Get(id)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		} else {
			klog.Infof("get app version %s failed, error: %s", id, err)
			return err
		}
	}

	switch appVersion.Status.State {
	case v1alpha1.StateActive:
		klog.Warningf("delete app version %s/%s not permitted, current state:%s", appVersion.GetWorkspace(),
			appVersion.GetTrueName(), appVersion.Status.State)
		return actionNotPermitted
	}

	// Delete data in storage
	err = c.backingStoreClient.Delete(dataKeyInStorage(appVersion.GetWorkspace(), id))
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != s3.ErrCodeNoSuchKey {
			klog.Errorf("delete app version %s/%s data failed, error: %s", appVersion.GetWorkspace(), appVersion.Name, err)
			return deleteDataInStorageFailed
		}
	}

	// delete app version in etcd
	err = c.appVersionClient.Delete(context.TODO(), id, metav1.DeleteOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("delete app version %s failed", err)
		return err
	} else {
		klog.Infof("app version %s deleted", id)
	}

	return nil
}

func (c *applicationOperator) DescribeAppVersion(id string) (*AppVersion, error) {
	version, err := c.versionLister.Get(id)
	if err != nil {
		klog.Errorf("get app version [%s] failed, error: %s", id, err)
		return nil, err
	}
	app := convertAppVersion(version)
	return app, nil
}

func (c *applicationOperator) ModifyAppVersion(id string, request *ModifyAppVersionRequest) error {

	version, err := c.versionLister.Get(id)
	if err != nil {
		klog.Errorf("get app version [%s] failed, error: %s", id, err)
		return err
	}

	versionCopy := version.DeepCopy()
	spec := &versionCopy.Spec
	if request.Name != nil && *request.Name != "" {
		spec.Version, spec.AppVersion = parseChartVersionName(*request.Name)
	}

	if request.Description != nil && *request.Description != "" {
		spec.Description = stringutils.ShortenString(*request.Description, v1alpha1.MsgLen)
	}
	patch := client.MergeFrom(version)
	data, err := patch.Data(versionCopy)
	if err != nil {
		klog.Error("create patch failed", err)
		return err
	}

	// data == "{}", need not to patch
	if len(data) == 2 {
		return nil
	}

	_, err = c.appVersionClient.Patch(context.TODO(), id, patch.Type(), data, metav1.PatchOptions{})

	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *applicationOperator) ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	versions, err := c.getAppVersionsByAppId(conditions.Match[AppId])

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	versions = filterAppVersions(versions, conditions)
	if reverse {
		sort.Sort(sort.Reverse(AppVersions(versions)))
	} else {
		sort.Sort(AppVersions(versions))
	}

	items := make([]interface{}, 0, int(math.Min(float64(limit), float64(len(versions)))))

	for i, j := offset, 0; i < len(versions) && j < limit; i, j = i+1, j+1 {
		items = append(items, convertAppVersion(versions[i]))
	}
	return &models.PageableResponse{Items: items, TotalCount: len(versions)}, nil
}

func (c *applicationOperator) ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	appVersions, err := c.versionLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	filtered := filterAppReviews(appVersions, conditions)
	if reverse {
		sort.Sort(sort.Reverse(AppVersionReviews(filtered)))
	} else {
		sort.Sort(AppVersionReviews(filtered))
	}

	items := make([]interface{}, 0, len(filtered))

	for i, j := offset, 0; i < len(filtered) && j < limit; i, j = i+1, j+1 {
		review := convertAppVersionReview(filtered[i])
		items = append(items, review)
	}

	return &models.PageableResponse{Items: items, TotalCount: len(filtered)}, nil
}

func (c *applicationOperator) ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	appId := conditions.Match[AppId]
	versionId := conditions.Match[VersionId]

	var versions []*v1alpha1.HelmApplicationVersion
	var err error
	if versionId == "" {
		ls := map[string]string{
			constants.ChartApplicationIdLabelKey: appId,
		}
		versions, err = c.versionLister.List(labels.SelectorFromSet(ls))
		if err != nil {
			klog.Errorf("get app %s failed, error: %s", appId, err)
		}
	} else {
		version, err := c.versionLister.Get(versionId)
		if err != nil {
			klog.Errorf("get app version %s failed, error: %s", versionId, err)
		}
		versions = []*v1alpha1.HelmApplicationVersion{version}
	}

	var allAudits []*AppVersionAudit
	for _, item := range versions {
		audits := convertAppVersionAudit(item)
		allAudits = append(allAudits, audits...)
	}

	sort.Sort(AppVersionAuditList(allAudits))

	items := make([]interface{}, 0, limit)

	for i, j := offset, 0; i < len(allAudits) && j < limit; i, j = i+1, j+1 {
		items = append(items, allAudits[i])
	}

	return &models.PageableResponse{Items: items, TotalCount: len(allAudits)}, nil
}

func (c *applicationOperator) DoAppVersionAction(versionId string, request *ActionRequest) error {
	var err error
	t := metav1.Now()
	var audit = v1alpha1.Audit{
		Message:  request.Message,
		Operator: request.Username,
		Time:     t,
	}
	state := v1alpha1.StateDraft

	version, err := c.versionLister.Get(versionId)
	if err != nil {
		klog.Errorf("get app version %s failed, error: %s", versionId, err)
		return err
	}

	switch request.Action {
	case ActionCancel:
		if version.Status.State != v1alpha1.StateSubmitted {
		}
		state = v1alpha1.StateDraft
		audit.State = v1alpha1.StateDraft
	case ActionPass:
		if version.Status.State != v1alpha1.StateSubmitted {

		}
		state = v1alpha1.StatePassed
		audit.State = v1alpha1.StatePassed
	case ActionRecover:
		if version.Status.State != v1alpha1.StateSuspended {

		}
		state = v1alpha1.StateActive
		audit.State = v1alpha1.StateActive
	case ActionReject:
		if version.Status.State != v1alpha1.StateSubmitted {
			// todo check status
		}
		state = v1alpha1.StateRejected
		audit.State = v1alpha1.StateRejected
	case ActionSubmit:
		if version.Status.State != v1alpha1.StateDraft {
			// todo check status
		}
		state = v1alpha1.StateSubmitted
		audit.State = v1alpha1.StateSubmitted
	case ActionSuspend:
		if version.Status.State != v1alpha1.StateActive {
			// todo check status
		}
		state = v1alpha1.StateSuspended
		audit.State = v1alpha1.StateSuspended
	case ActionRelease:
		// release to app store
		if version.Status.State != v1alpha1.StatePassed {
			// todo check status
		}
		state = v1alpha1.StateActive
		audit.State = v1alpha1.StateActive
	default:
		err = errors.New("action not support")
	}

	_ = state
	if err != nil {
		klog.Error(err)
		return err
	}

	version, err = c.updateAppVersionStatus(version, state, &audit)

	if err != nil {
		klog.Errorf("update app version audit [%s] failed, error: %s", versionId, err)
		return err
	}

	if request.Action == ActionRelease || request.Action == ActionRecover {
		// if we release a new helm application version, we need update the spec in helm application copy
		app, err := c.appLister.Get(version.GetHelmApplicationId())
		if err != nil {
			return err
		}
		appInStore, err := c.appLister.Get(fmt.Sprintf("%s%s", version.GetHelmApplicationId(), v1alpha1.HelmApplicationAppStoreSuffix))
		if err != nil {
			if apierrors.IsNotFound(err) {
				// controller-manager will create application in app store
				return nil
			}
			return err
		}

		if !reflect.DeepEqual(&app.Spec, &appInStore.Spec) {
			appCopy := appInStore.DeepCopy()
			appCopy.Spec = app.Spec
			patch := client.MergeFrom(appInStore)
			data, _ := patch.Data(appCopy)
			_, err = c.appClient.Patch(context.TODO(), appCopy.Name, patch.Type(), data, metav1.PatchOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Create helmApplicationVersion and helmAudit
func (c *applicationOperator) createApplicationVersion(ver *v1alpha1.HelmApplicationVersion) (*v1alpha1.HelmApplicationVersion, error) {
	ls := map[string]string{
		constants.ChartApplicationIdLabelKey: ver.GetHelmApplicationId(),
	}

	list, err := c.versionLister.List(labels.SelectorFromSet(ls))

	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if len(list) > 0 {
		verName := ver.GetVersionName()
		for _, v := range list {
			if verName == v.GetVersionName() {
				klog.V(2).Infof("helm application version: %s exist", verName)
				return nil, appVersionItemExists
			}
		}
	}

	// save chart data to s3 storage
	_, err = base64.StdEncoding.Decode(ver.Spec.Data, ver.Spec.Data)
	if err != nil {
		klog.Errorf("decode error: %s", err)
		return nil, err
	} else {
		err = c.backingStoreClient.Upload(dataKeyInStorage(ver.GetWorkspace(), ver.Name), ver.Name, bytes.NewReader(ver.Spec.Data))
		if err != nil {
			klog.Errorf("upload chart for app version: %s/%s failed, error: %s", ver.GetWorkspace(),
				ver.GetTrueName(), err)
			return nil, uploadChartDataFailed
		} else {
			klog.V(4).Infof("chart data uploaded for app version: %s/%s", ver.GetWorkspace(), ver.GetTrueName())
		}
	}

	// data will not save to etcd
	ver.Spec.Data = nil
	ver.Spec.DataKey = ver.Name
	version, err := c.appVersionClient.Create(context.TODO(), ver, metav1.CreateOptions{})
	if err == nil {
		klog.V(4).Infof("create helm application %s version success", version.Name)
	}

	return version, err
}

func (c *applicationOperator) updateAppVersionStatus(version *v1alpha1.HelmApplicationVersion, state string, status *v1alpha1.Audit) (*v1alpha1.HelmApplicationVersion, error) {
	version.Status.State = state

	states := append([]v1alpha1.Audit{*status}, version.Status.Audit...)
	if len(version.Status.Audit) >= v1alpha1.HelmRepoSyncStateLen {
		// strip the last item
		states = states[:v1alpha1.HelmRepoSyncStateLen:v1alpha1.HelmRepoSyncStateLen]
	}

	version.Status.Audit = states
	version, err := c.appVersionClient.UpdateStatus(context.TODO(), version, metav1.UpdateOptions{})

	return version, err
}

func (c *applicationOperator) GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error) {
	var version *v1alpha1.HelmApplicationVersion
	var err error

	version, err = c.getAppVersionByVersionIdWithData(versionId)
	if err != nil {
		return nil, err
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(version.Spec.Data))
	if err != nil {
		klog.Errorf("read app version %s failed, error: %s", versionId, err)
		return nil, err
	}

	tarReader := tar.NewReader(gzReader)

	res := &GetAppVersionPackageFilesResponse{Files: map[string]strfmt.Base64{}, VersionId: versionId}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			klog.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
			return res, err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			curData, _ := ioutil.ReadAll(tarReader)
			name := strings.TrimLeft(header.Name, fmt.Sprintf("%s/", version.GetTrueName()))
			res.Files[name] = curData
		default:
			klog.Errorf(
				"ExtractTarGz: unknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return res, nil
}

func (c *applicationOperator) getAppVersionByVersionIdWithData(versionId string) (*v1alpha1.HelmApplicationVersion, error) {
	if version, exists, err := c.cachedRepos.GetAppVersionWithData(versionId); exists {
		if err != nil {
			return nil, err
		}
		return version, nil
	}

	version, err := c.versionLister.Get(versionId)
	if err != nil {
		return nil, err
	}

	data, err := c.backingStoreClient.Read(dataKeyInStorage(version.GetWorkspace(), versionId))
	if err != nil {
		klog.Errorf("load chart data for app version: %s/%s failed, error : %s", version.GetTrueName(),
			version.GetTrueName(), err)
		return nil, downloadFileFailed
	}
	version.Spec.Data = data

	return version, nil
}

func (c *applicationOperator) getAppVersionsByAppId(appId string) (ret []*v1alpha1.HelmApplicationVersion, err error) {
	if ret, exists := c.cachedRepos.ListAppVersionsByAppId(appId); exists {
		return ret, nil
	}

	// list app version from client-go
	ret, err = c.versionLister.List(labels.SelectorFromSet(map[string]string{constants.ChartApplicationIdLabelKey: appId}))
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	return
}
