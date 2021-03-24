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
	"bytes"
	"context"
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	v1alpha13 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/reposcache"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"time"
)

type ApplicationInterface interface {
	ListApps(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	DescribeApp(id string) (*App, error)
	DeleteApp(id string) error
	CreateApp(req *CreateAppRequest) (*CreateAppResponse, error)
	ModifyApp(appId string, request *ModifyAppRequest) error
	DeleteAppVersion(id string) error
	ModifyAppVersion(id string, request *ModifyAppVersionRequest) error
	DescribeAppVersion(id string) (*AppVersion, error)
	CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error)
	ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error)
	GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error)
	DoAppAction(appId string, request *ActionRequest) error
	DoAppVersionAction(versionId string, request *ActionRequest) error
	ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error)
	ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
}

type applicationOperator struct {
	backingStoreClient s3.Interface
	informers          externalversions.SharedInformerFactory

	appClient        v1alpha13.HelmApplicationInterface
	appVersionClient v1alpha13.HelmApplicationVersionInterface

	appLister     listers_v1alpha1.HelmApplicationLister
	versionLister listers_v1alpha1.HelmApplicationVersionLister

	repoLister listers_v1alpha1.HelmRepoLister
	ctgLister  listers_v1alpha1.HelmCategoryLister
	rlsLister  listers_v1alpha1.HelmReleaseLister

	cachedRepos reposcache.ReposCache
}

func newApplicationOperator(cached reposcache.ReposCache, informers externalversions.SharedInformerFactory, ksClient versioned.Interface, storeClient s3.Interface) ApplicationInterface {
	op := &applicationOperator{
		backingStoreClient: storeClient,
		informers:          informers,
		repoLister:         informers.Application().V1alpha1().HelmRepos().Lister(),

		appClient:        ksClient.ApplicationV1alpha1().HelmApplications(),
		appVersionClient: ksClient.ApplicationV1alpha1().HelmApplicationVersions(),

		appLister:     informers.Application().V1alpha1().HelmApplications().Lister(),
		versionLister: informers.Application().V1alpha1().HelmApplicationVersions().Lister(),

		ctgLister:   informers.Application().V1alpha1().HelmCategories().Lister(),
		rlsLister:   informers.Application().V1alpha1().HelmReleases().Lister(),
		cachedRepos: cached,
	}

	return op
}

// save icon data and helm application
func (c *applicationOperator) createApp(app *v1alpha1.HelmApplication, iconData []byte) (*v1alpha1.HelmApplication, error) {
	exists, err := c.getHelmAppByName(app.GetWorkspace(), app.GetTrueName())
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, appItemExists
	}

	if len(iconData) != 0 {
		// save icon attachment
		iconId := idutils.GetUuid(v1alpha1.HelmAttachmentPrefix)
		err = c.backingStoreClient.Upload(iconId, iconId, bytes.NewBuffer(iconData))
		if err != nil {
			klog.Errorf("save icon attachment failed, error: %s", err)
			return nil, err
		}
		app.Spec.Icon = iconId
	}

	app, err = c.appClient.Create(context.TODO(), app, metav1.CreateOptions{})
	return app, err
}

// get helm app by name in workspace
func (c *applicationOperator) getHelmAppByName(workspace, name string) (*v1alpha1.HelmApplication, error) {
	ls := map[string]string{
		constants.WorkspaceLabelKey: workspace,
	}

	list, err := c.appLister.List(labels.SelectorFromSet(ls))

	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if len(list) > 0 {
		for _, a := range list {
			if a.GetTrueName() == name {
				return a, nil
			}
		}
	}

	return nil, nil
}

func (c *applicationOperator) ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error) {

	chrt, err := helmrepoindex.LoadPackage(request.VersionPackage)

	result := &ValidatePackageResponse{}

	if err != nil {
		matchPackageFailedError(err, result)
		if (result.Error == "EOF" || result.Error == "") && len(result.ErrorDetails) == 0 {
			klog.Errorf("package parse failed, error: %s", err.Error())
			return nil, errors.New("package parse failed")
		}
	} else {
		result.Name = chrt.GetName()
		result.VersionName = chrt.GetVersionName()
		result.Description = chrt.GetDescription()
		result.URL = chrt.GetUrls()
	}

	return result, nil
}

func (c *applicationOperator) DoAppAction(appId string, request *ActionRequest) error {

	app, err := c.appLister.Get(appId)
	if err != nil {
		return err
	}

	var filterState string
	switch request.Action {
	case ActionSuspend:
		if app.Status.State != v1alpha1.StateActive {
			err = actionNotSupport
		}
		filterState = v1alpha1.StateActive
	case ActionRecover:
		if app.Status.State != v1alpha1.StateSuspended {
			err = actionNotSupport
		}
		filterState = v1alpha1.StateSuspended
	default:
		err = actionNotSupport
	}
	if err != nil {
		return err
	}

	var versions []*v1alpha1.HelmApplicationVersion
	ls := map[string]string{
		constants.ChartApplicationIdLabelKey: appId,
	}
	versions, err = c.versionLister.List(labels.SelectorFromSet(ls))
	if err != nil {
		klog.Errorf("get helm app %s version failed, error: %s", appId, err)
		return err
	}

	versions = filterAppVersionByState(versions, []string{filterState})
	for _, version := range versions {
		err = c.DoAppVersionAction(version.GetHelmApplicationVersionId(), request)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *applicationOperator) CreateApp(req *CreateAppRequest) (*CreateAppResponse, error) {
	if c.backingStoreClient == nil {
		return nil, invalidS3Config
	}
	chrt, err := helmrepoindex.LoadPackage(req.VersionPackage)
	if err != nil {
		klog.Errorf("load package %s/%s failed, error: %s", req.Isv, req.Name, err)
		return nil, err
	}

	// create helm application
	name := idutils.GetUuid36(v1alpha1.HelmApplicationIdPrefix)
	helmApp := &v1alpha1.HelmApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: req.Username,
			},
			Labels: map[string]string{
				constants.WorkspaceLabelKey: req.Isv,
			},
		},
		Spec: v1alpha1.HelmApplicationSpec{
			Name:        req.Name,
			Description: stringutils.ShortenString(chrt.GetDescription(), v1alpha1.MsgLen),
		},
	}
	app, err := c.createApp(helmApp, req.Icon)
	if err != nil {
		klog.Errorf("create helm application %s/%s failed, error: %s", req.Isv, req.Name, err)
		if helmApp.Spec.Icon != "" {
			c.backingStoreClient.Delete(helmApp.Spec.Icon)
		}
		return nil, err
	} else {
		klog.V(4).Infof("helm application %s/%s created, app id: %s", req.Isv, req.Name, app.Name)
	}

	// create app version
	chartPackage := req.VersionPackage.String()
	ver := buildApplicationVersion(app, chrt, &chartPackage, req.Username)
	ver, err = c.createApplicationVersion(ver)

	if err != nil {
		klog.Errorf("create helm application %s/%s versions failed, error: %s", req.Isv, req.Name, err)
		return nil, err
	} else {
		klog.V(4).Infof("helm application version %s/%s created, app version id: %s", req.Isv, req.Name, ver.Name)
	}

	return &CreateAppResponse{
		AppID:     app.GetHelmApplicationId(),
		VersionID: ver.GetHelmApplicationVersionId(),
	}, nil
}

func buildLabelSelector(conditions *params.Conditions) map[string]string {
	ls := make(map[string]string)

	repoId := conditions.Match[RepoId]
	// app store come first
	if repoId != "" {
		ls[constants.ChartRepoIdLabelKey] = repoId
	} else {
		if conditions.Match[WorkspaceLabel] != "" {
			ls[constants.WorkspaceLabelKey] = conditions.Match[WorkspaceLabel]
		}
	}
	if conditions.Match[CategoryId] != "" {
		ls[constants.CategoryIdLabelKey] = conditions.Match[CategoryId]
	}

	return ls
}

func (c *applicationOperator) ListApps(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	apps, err := c.listApps(conditions)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	apps = filterApps(apps, conditions)

	if reverse {
		sort.Sort(sort.Reverse(HelmApplicationList(apps)))
	} else {
		sort.Sort(HelmApplicationList(apps))
	}

	items := make([]interface{}, 0, limit)

	for i, j := offset, 0; i < len(apps) && j < limit; i, j = i+1, j+1 {
		versions, err := c.getAppVersionsByAppId(apps[i].GetHelmApplicationId())
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, err
		}

		ctg, _ := c.ctgLister.Get(apps[i].GetHelmCategoryId())

		items = append(items, convertApp(apps[i], versions, ctg, 0))
	}
	return &models.PageableResponse{Items: items, TotalCount: len(apps)}, nil
}

func (c *applicationOperator) DeleteApp(id string) error {
	app, err := c.appLister.Get(id)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		} else {
			klog.Errorf("get app %s failed, error: %s", id, err)
			return err
		}
	}

	ls := map[string]string{
		constants.ChartApplicationIdLabelKey: app.GetHelmApplicationId(),
	}

	list, err := c.versionLister.List(labels.SelectorFromSet(ls))
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("versions of app %s has been deleted", id)
		} else {
			klog.Error(err)
			return err
		}
	} else if len(list) > 0 {
		return fmt.Errorf("app %s has some versions not deleted", id)
	}

	err = c.appClient.Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("delete app %s failed, error: %s", id, err)
		return err
	} else {
		c.deleteAppAttachment(app)
		klog.V(4).Infof("app %s deleted", app.Name)
	}

	// delete application in app store
	id = fmt.Sprintf("%s%s", id, v1alpha1.HelmApplicationAppStoreSuffix)
	app, err = c.appClient.Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		} else {
			klog.Errorf("get app %s failed, error: %s", id, err)
			return err
		}
	}

	// delete application in app store
	err = c.appClient.Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("delete app %s failed, error: %s", id, err)
		return err
	} else {
		c.deleteAppAttachment(app)
		klog.V(4).Infof("app %s deleted", app.Name)
	}

	return nil
}

func (c *applicationOperator) ModifyApp(appId string, request *ModifyAppRequest) error {
	if c.backingStoreClient == nil {
		return invalidS3Config
	}

	app, err := c.appLister.Get(appId)
	if err != nil {
		klog.Error(err)
		return err
	}

	appCopy := app.DeepCopy()
	// modify category
	if request.CategoryID != nil {
		if *request.CategoryID == "" {
			delete(appCopy.Labels, constants.CategoryIdLabelKey)
			klog.V(4).Infof("delete app %s category", app.Name)
		} else {
			appCopy.Labels[constants.CategoryIdLabelKey] = *request.CategoryID
			klog.V(4).Infof("set app %s category to %s", app.Name, *request.CategoryID)
		}
	}

	// modify app name
	if request.Name != nil && len(*request.Name) > 0 && app.GetTrueName() != *request.Name {
		existsApp, err := c.getHelmAppByName(app.GetWorkspace(), *request.Name)
		if err != nil {
			return err
		}
		if existsApp != nil {
			return appItemExists
		}
		klog.V(4).Infof("change app %s name from %s to %s", app.Name, app.GetTrueName(), *request.Name)
		appCopy.Spec.Name = *request.Name
	}

	// save app attachment and icon
	add, err := c.modifyAppAttachment(appCopy, request)
	if err != nil {
		klog.Errorf("add app attachment %s failed, error: %s", appCopy.Name, err)
		return err
	}

	if request.Description != nil {
		appCopy.Spec.Description = *request.Description
	}
	if request.Abstraction != nil {
		appCopy.Spec.Abstraction = *request.Abstraction
	}

	if request.Home != nil {
		appCopy.Spec.AppHome = *request.Home
	}
	appCopy.Status.UpdateTime = &metav1.Time{Time: time.Now()}

	patch := client.MergeFrom(app)
	data, err := patch.Data(appCopy)
	if err != nil {
		klog.Errorf("create patch failed, error: %s", err)
		return err
	}

	_, err = c.appClient.Patch(context.TODO(), appId, patch.Type(), data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("patch helm application: %s failed, error: %s", appId, err)
		if add != "" {
			// if patch failed, delete saved icon or attachment
			c.backingStoreClient.Delete(add)
		}
		return err
	}
	return nil
}

func (c *applicationOperator) deleteAppAttachment(app *v1alpha1.HelmApplication) {
	if app.Spec.Icon != "" {
		c.backingStoreClient.Delete(app.Spec.Icon)
	}

	for _, id := range app.Spec.Attachments {
		c.backingStoreClient.Delete(id)
	}
}

func (c *applicationOperator) modifyAppAttachment(app *v1alpha1.HelmApplication, request *ModifyAppRequest) (add string, err error) {
	if request.Type == nil {
		return "", nil
	}
	switch *request.Type {
	case v1alpha1.AttachmentTypeScreenshot:
		if request.Sequence == nil {
			return "", nil
		}
		seq := *request.Sequence
		attachments := &app.Spec.Attachments
		if len(request.AttachmentContent) == 0 {
			// delete old attachments
			if len(*attachments) > int(seq) {
				del := (*attachments)[seq]
				err = c.backingStoreClient.Delete(del)
				if err != nil {
					return "", err
				} else {
					*attachments = append((*attachments)[:seq], (*attachments)[seq+1:]...)
				}
			}
		} else {
			if len(*attachments) < 6 {
				// add attachment to app
				add := idutils.GetUuid("att-")
				*attachments = append(*attachments, add)
				err = c.backingStoreClient.Upload(add, add, bytes.NewBuffer(request.AttachmentContent))
				if err != nil {
					return "", err
				} else {
					return add, nil
				}
			}
		}
	case v1alpha1.AttachmentTypeIcon: // modify app icon
		// delete old icon
		if app.Spec.Icon != "" {
			err = c.backingStoreClient.Delete(app.Spec.Icon)
			if err != nil {
				return "", err
			}
		}
	}
	if len(request.AttachmentContent) != 0 {
		add := idutils.GetUuid("att-")
		err = c.backingStoreClient.Upload(add, add, bytes.NewBuffer(request.AttachmentContent))
		if err != nil {
			return "", err
		} else {
			app.Spec.Icon = add
			return add, nil
		}
	}
	return "", nil
}

// modify icon or attachment of the app
// added: new attachments have been saved to store
// deleted: attachments should be deleted
func (c *applicationOperator) appAttachmentDiff(old, newApp *v1alpha1.HelmApplication) (added, deleted []string) {

	added = make([]string, 0, 7)
	deleted = make([]string, 0, 7)

	if old.Spec.Icon != newApp.Spec.Icon {
		if old.Spec.Icon != "" && !strings.HasPrefix(old.Spec.Icon, "http://") {
			deleted = append(deleted, old.Spec.Icon)
		}
		added = append(added, newApp.Spec.Icon)
	}

	existsAtt := make(map[string]string, 6)
	newAtt := make(map[string]string, 6)

	for _, id := range newApp.Spec.Attachments {
		newAtt[id] = ""
	}

	for _, id := range old.Spec.Attachments {
		existsAtt[id] = ""
	}

	for _, id := range newApp.Spec.Attachments {
		if _, exists := existsAtt[id]; !exists {
			added = append(added, id)
		}
	}

	for _, id := range old.Spec.Attachments {
		if _, exists := newAtt[id]; !exists {
			deleted = append(deleted, id)
		}
	}

	return added, deleted
}

func (c *applicationOperator) DescribeApp(id string) (*App, error) {
	var helmApp *v1alpha1.HelmApplication
	var ctg *v1alpha1.HelmCategory
	var err error

	helmApp, err = c.getHelmApplication(id)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	versions, err := c.getAppVersionsByAppId(helmApp.GetHelmApplicationId())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	ctg, err = c.ctgLister.Get(helmApp.GetHelmCategoryId())

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	app := convertApp(helmApp, versions, ctg, 0)
	return app, nil
}

func (c *applicationOperator) listApps(conditions *params.Conditions) (ret []*v1alpha1.HelmApplication, err error) {
	repoId := conditions.Match[RepoId]
	if repoId != "" && repoId != v1alpha1.AppStoreRepoId {
		// get helm application from helm repo
		if ret, exists := c.cachedRepos.ListApplicationsByRepoId(repoId); !exists {
			klog.Warningf("load repo failed, repo id: %s", repoId)
			return nil, loadRepoInfoFailed
		} else {
			return ret, nil
		}
	} else {
		if c.backingStoreClient == nil {
			return []*v1alpha1.HelmApplication{}, nil
		}
		ret, err = c.appLister.List(labels.SelectorFromSet(buildLabelSelector(conditions)))
	}

	return
}

func (c *applicationOperator) getHelmApplication(appId string) (*v1alpha1.HelmApplication, error) {
	if app, exists := c.cachedRepos.GetApplication(appId); exists {
		return app, nil
	} else {
		return c.appLister.Get(appId)
	}
}
