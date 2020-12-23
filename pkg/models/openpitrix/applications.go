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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	v1alpha12 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	v1beta12 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	v1beta13 "kubesphere.io/kubesphere/pkg/client/listers/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/helpers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	strings2 "strings"
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
	//GetAppVersionPackage(appId, versionId string) (*GetAppVersionPackageResponse, error)
	DoAppAction(appId string, request *ActionRequest) error
	DoAppVersionAction(versionId string, request *ActionRequest) error
	ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error)
	ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
}

type applicationOperator struct {
	informers        externalversions.SharedInformerFactory
	appClient        v1beta12.FederatedHelmApplicationInterface
	appVersionClient v1beta12.FederatedHelmApplicationVersionInterface
	auditClient      v1alpha12.HelmAuditInterface

	auditLister   listers_v1alpha1.HelmAuditLister
	appLister     v1beta13.FederatedHelmApplicationLister
	versionLister v1beta13.FederatedHelmApplicationVersionLister

	repoLister listers_v1alpha1.HelmRepoLister
	ctgLister  listers_v1alpha1.HelmCategoryLister
	rlsLister  listers_v1alpha1.HelmReleaseLister

	useFederatedResource bool
	cachedRepos          *cachedRepos
}

func newApplicationOperator(cached *cachedRepos, informers externalversions.SharedInformerFactory, ksClient versioned.Interface, useFederatedResource bool) ApplicationInterface {
	op := &applicationOperator{
		useFederatedResource: useFederatedResource,
		informers:            informers,
		repoLister:           informers.Application().V1alpha1().HelmRepos().Lister(),
		auditClient:          ksClient.ApplicationV1alpha1().HelmAudits(),

		auditLister: informers.Application().V1alpha1().HelmAudits().Lister(),

		ctgLister:   informers.Application().V1alpha1().HelmCategories().Lister(),
		rlsLister:   informers.Application().V1alpha1().HelmReleases().Lister(),
		cachedRepos: cached,
	}

	if op.useFederatedResource {
		op.appClient = ksClient.TypesV1beta1().FederatedHelmApplications()
		op.appVersionClient = ksClient.TypesV1beta1().FederatedHelmApplicationVersions()
		op.appLister = informers.Types().V1beta1().FederatedHelmApplications().Lister()
		op.versionLister = informers.Types().V1beta1().FederatedHelmApplicationVersions().Lister()
	}

	return op
}

//get helm app by name in workspace
func (c *applicationOperator) getHelmAppByName(workspace, name string) (*v1beta1.FederatedHelmApplication, error) {
	ls := map[string]string{
		constants.WorkspaceLabelKey: workspace,
	}

	list, err := c.appLister.List(helpers.MapAsLabelSelector(ls))

	if err != nil && !apiErrors.IsNotFound(err) {
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
func (c *applicationOperator) createApp(app *v1beta1.FederatedHelmApplication) (*v1beta1.FederatedHelmApplication, error) {
	exists, err := c.getHelmAppByName(app.GetWorkspace(), app.GetTrueName())
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, ItemExists
	}
	app, err = c.appClient.Create(context.TODO(), app, metav1.CreateOptions{})
	return app, err
}

func (c *applicationOperator) ValidatePackage(request *ValidatePackageRequest) (*ValidatePackageResponse, error) {

	chrt, err := helmrepoindex.LoadPackage(request.VersionPackage)

	result := &ValidatePackageResponse{}

	if err != nil {
		matchPackageFailedError(err, result)
		if result.Error == "" && len(result.ErrorDetails) == 0 {
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

	template := app.Spec.Template
	var filterState string
	switch request.Action {
	case ActionSuspend:
		if template.Spec.Status != constants.StateActive {
			err = errors.New("action not support")
		}
		filterState = constants.StateActive
	case ActionRecover:
		if template.Spec.Status != constants.StateSuspended {
			err = errors.New("action not support")
		}
		filterState = constants.StateSuspended
	default:
		err = errors.New("action not support")
	}
	if err != nil {
		return err
	}

	var versions []*v1beta1.FederatedHelmApplicationVersion
	ls := map[string]string{
		constants.ChartApplicationIdLabelKey: appId,
	}
	versions, err = c.versionLister.List(helpers.MapAsLabelSelector(ls))
	if err != nil {
		klog.Errorf("get helm app %s version failed, error: %s", appId, err)
		return err
	}

	for _, version := range versions {
		version, err = c.fillAppVersionAudit(version)
		if err != nil {
			klog.Errorf("get helm app version audit %s failed, error: %s", version.Name, err)
			return err
		}
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

	chrt, err := helmrepoindex.LoadPackage(req.VersionPackage)
	if err != nil {
		return nil, err
	}

	iconId := ""

	if len(req.Icon) != 0 {
		//save icon attachment
		iconId = idutils.GetUuid(constants.HelmAttachmentPrefix)
		err = attachmentHandler.Save(iconId, bytes.NewBuffer(req.Icon))
		if err != nil {
			klog.Errorf("save icon attachment failed, error: %s", err)
			return nil, err
		}
	}

	//create helm application resource
	name := idutils.GetUuid36(constants.HelmApplicationIdPrefix)

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
			Icon:        iconId,
			Status:      constants.StateDraft,
			Description: strings.ShortenString(chrt.GetDescription(), constants.MsgLen),
		},
	}
	app, err := c.createApp(toFederateHelmApplication(helmApp))
	if err != nil {
		attachmentHandler.Delete(iconId)
		return nil, err
	}
	chartPackage := req.VersionPackage.String()
	ver := buildApplicationVersion(app, chrt, &chartPackage, req.Username)

	ver, err = c.createApplicationVersionWithAudit(ver)

	if err != nil {
		attachmentHandler.Delete(iconId)
		klog.Error(err)
		return nil, err
	}

	return &CreateAppResponse{
		AppID:     app.GetHelmApplicationId(),
		VersionID: ver.GetHelmApplicationVersionId(),
	}, nil
}

func buildLabelSelector(conditions *params.Conditions) map[string]string {
	ls := make(map[string]string)

	repoId := conditions.Match[RepoId]
	//app store come first
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
	if conditions.Match[Keyword] != "" {
		apps = helmApplicationFilter(conditions.Match[Keyword], apps)
	}

	if reverse {
		sort.Sort(sort.Reverse(FederatedHelmApplicationList(apps)))
	} else {
		sort.Sort(FederatedHelmApplicationList(apps))
	}

	items := make([]interface{}, 0, limit)

	for i, j := offset, 0; i < len(apps) && j < limit; {
		versions, err := c.listAppVersions(apps[i].GetHelmApplicationId())
		if err != nil && !apiErrors.IsNotFound(err) {
			return nil, err
		}

		rls, _ := c.rlsLister.List(helpers.MapAsLabelSelector(map[string]string{
			constants.ChartApplicationIdLabelKey: apps[i].GetHelmApplicationId(),
		}))
		ctg, _ := c.ctgLister.Get(apps[i].GetHelmCategoryId())

		items = append(items, convertApp(apps[i], versions, ctg, len(rls)))
		i++
		j++
	}
	return &models.PageableResponse{Items: items, TotalCount: len(apps)}, nil
}

func (c *applicationOperator) DeleteApp(id string) error {
	app, err := c.appLister.Get(id)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			klog.V(4).Infof("app %s has been deleted", id)
			return nil
		} else {
			klog.Error(err)
			return nil
		}
	}

	ls := map[string]string{
		constants.ChartApplicationIdLabelKey: app.GetHelmApplicationId(),
	}
	releases, err := c.rlsLister.List(helpers.MapAsLabelSelector(ls))

	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Error(err)
		return err
	} else if len(releases) > 0 {
		return fmt.Errorf("app %s has some releases not deleted", id)
	}

	list, err := c.versionLister.List(helpers.MapAsLabelSelector(ls))
	if err != nil {
		if apiErrors.IsNotFound(err) {
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
	}

	//delete application in app store
	id = fmt.Sprintf("%s%s", id, constants.HelmApplicationAppStoreSuffix)
	err = c.appClient.Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Errorf("delete app %s failed, error: %s", id, err)
		return err
	}

	return nil
}

func (c *applicationOperator) ModifyApp(appId string, request *ModifyAppRequest) error {
	app, err := c.appLister.Get(appId)
	if err != nil {
		klog.Error(err)
		return err
	}

	appCopy := app.DeepCopy()
	if request.CategoryID != nil {
		if *request.CategoryID == "" {
			delete(appCopy.Labels, constants.CategoryIdLabelKey)
		} else {
			appCopy.Labels[constants.CategoryIdLabelKey] = *request.CategoryID
		}
	}

	if request.Name != nil && app.GetTrueName() != *request.Name {
		existsApp, err := c.getHelmAppByName(app.GetWorkspace(), *request.Name)
		if err != nil {
			return err
		}
		if existsApp != nil {
			return ItemExists
		}
		if *request.Name != "" {
			//appCopy.Labels[constants.NameLabelKey] = *request.Name
			appCopy.Spec.Template.Spec.Name = *request.Name
		}
	}

	//save app attachment and icon
	add, err := appAttachment(appCopy, request)
	if err != nil {
		klog.Errorf("add app attachment %s failed, error: %s", appCopy.Name, err)
		return err
	}

	if request.Description != nil {
		appCopy.Spec.Template.Spec.Description = *request.Description
	}
	if request.Abstraction != nil {
		appCopy.Spec.Template.Spec.Abstraction = *request.Abstraction
	}

	if request.Home != nil {
		appCopy.Spec.Template.Spec.AppHome = *request.Home
	}
	appCopy.Spec.Template.Spec.UpdateTime = &metav1.Time{Time: time.Now()}

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
			attachmentHandler.Delete(add)
		}
		return err
	}
	return nil
}

func appAttachment(app *v1beta1.FederatedHelmApplication, request *ModifyAppRequest) (add string, err error) {
	if request.Type == nil {
		return "", nil
	}

	switch *request.Type {
	case constants.AttachmentTypeScreenshot:
		if request.Sequence == nil {
			return "", nil
		}
		seq := *request.Sequence
		attachments := &app.Spec.Template.Spec.Attachments
		if len(request.AttachmentContent) == 0 {
			//delete attachment from app
			if len(*attachments) > int(seq) {
				del := (*attachments)[seq]
				err = attachmentHandler.Delete(del)
				if err != nil {
					return "", err
				} else {
					*attachments = append((*attachments)[:seq], (*attachments)[seq+1:]...)
				}
			}
		} else {
			if len(*attachments) < 6 {
				//add attachment to app
				add := idutils.GetUuid("att-")
				*attachments = append(*attachments, add)
				err = attachmentHandler.Save(add, bytes.NewBuffer(request.AttachmentContent))
				if err != nil {
					return "", err
				} else {
					return add, nil
				}
			}
		}
	case constants.AttachmentTypeIcon:
		if app.Spec.Template.Spec.Icon != "" {
			err = attachmentHandler.Delete(app.Spec.Template.Spec.Icon)
			if err != nil {
				return "", err
			}
		}
		if len(request.AttachmentContent) != 0 {
			add := idutils.GetUuid("att-")
			err = attachmentHandler.Save(add, bytes.NewBuffer(request.AttachmentContent))
			if err != nil {
				return "", err
			} else {
				app.Spec.Template.Spec.Icon = add
				return add, nil
			}
		}
	}

	return "", nil
}

func (c *applicationOperator) DescribeApp(id string) (*App, error) {
	var helmApp *v1beta1.FederatedHelmApplication
	var ctg *v1alpha1.HelmCategory
	var err error

	helmApp, err = c.getHelmApplication(id)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	versions, err := c.listAppVersions(helmApp.GetHelmApplicationId())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	ctg, err = c.ctgLister.Get(helmApp.GetHelmCategoryId())

	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	app := convertApp(helmApp, versions, ctg, 0)
	return app, nil
}

func helmApplicationFilter(namePrefix string, list []*v1beta1.FederatedHelmApplication) (res []*v1beta1.FederatedHelmApplication) {
	for _, repo := range list {
		name := repo.GetTrueName()
		if strings2.HasPrefix(name, namePrefix) {
			res = append(res, repo)
		}
	}
	return
}

func (c *applicationOperator) listApps(conditions *params.Conditions) (ret []*v1beta1.FederatedHelmApplication, err error) {
	repoId := conditions.Match[RepoId]
	if repoId != "" && repoId != constants.AppStoreRepoId {
		//get helm application from helm repo
		c.cachedRepos.RLock()
		defer c.cachedRepos.RUnlock()
		cached := c.cachedRepos
		repo := cached.repos[repoId]
		if repo == nil {
			return ret, nil
		}

		ret = make([]*v1beta1.FederatedHelmApplication, 0, 10)
		for _, a := range cached.apps {
			if a.GetHelmRepoId() == repo.Name {
				ret = append(ret, a)
			}
		}

		return ret, nil
	} else {
		ret, err = c.appLister.List(helpers.MapAsLabelSelector(buildLabelSelector(conditions)))
		var states []string
		if conditions.Match[Status] != "" {
			states = strings2.Split(conditions.Match[Status], "|")
		}
		ret = filterHelmApplicationByStates(ret, states)
	}

	return
}

func (c *applicationOperator) getHelmApplication(appId string) (*v1beta1.FederatedHelmApplication, error) {
	c.cachedRepos.RLock()
	cache := c.cachedRepos
	if app, exists := cache.apps[appId]; exists {
		c.cachedRepos.RUnlock()
		return app, nil
	} else {
		c.cachedRepos.RUnlock()

		return c.appLister.Get(appId)
	}
}

func filterHelmApplicationByStates(apps []*v1beta1.FederatedHelmApplication, states []string) []*v1beta1.FederatedHelmApplication {

	if len(states) == 0 || len(apps) == 0 {
		return apps
	}

	var j = 0
	for i := 0; i < len(apps); i++ {
		//state := apps[i].Status.State
		//TODO, app state
		state := apps[i].Spec.Template.Spec.Status
		//default value is draft
		if state == "" {
			state = constants.StateDraft
		}
		if sliceutil.HasString(states, state) {
			if i != j {
				apps[j] = apps[i]
			}
			j++
		}
	}

	apps = apps[:j:j]
	return apps
}

//func (c *applicationOperator) getAppVersion(id string) (ret *v1alpha1.HelmApplicationVersion, err error) {
//	cached := c.cachedRepos.Load().(cachedRepos)
//	ret = cached.versions[id]
//	if ret != nil {
//		return ret, nil
//	}
//	ret, err = c.versionLister.Get(id)
//	if err != nil && !apiErrors.IsNotFound(err) {
//		klog.Error(err)
//		return nil, err
//	}
//
//	return
//}
