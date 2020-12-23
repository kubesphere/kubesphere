package openpitrix

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	"io"
	"io/ioutil"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/helpers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"math"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	strings2 "strings"
)

func (c *applicationOperator) CreateAppVersion(request *CreateAppVersionRequest) (*CreateAppVersionResponse, error) {

	chrt, err := helmrepoindex.LoadPackage(request.Package)
	if err != nil {
		return nil, err
	}

	app, err := c.appLister.Get(request.AppId)

	if err != nil {
		klog.Errorf("get app %s failed, error: %s", request.AppId, err)
		return nil, err
	}
	chartPackage := request.Package.String()
	version := buildApplicationVersion(app, chrt, &chartPackage, request.Username)
	version, err = c.createApplicationVersionWithAudit(version)

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

	//err := c.appClient.HelmApplicationVersions().Delete(context.TODO(), id, metav1.DeleteOptions{})
	err := c.appVersionClient.Delete(context.TODO(), id, metav1.DeleteOptions{})

	if err != nil {
		if apiErrors.IsNotFound(err) {
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
	template := &versionCopy.Spec.Template
	if request.Name != nil && *request.Name != "" {
		template.Spec.Version, template.Spec.AppVersion = parseChartVersionName(*request.Name)
	}

	if request.Description != nil && *request.Description != "" {
		template.Spec.Description = strings.ShortenString(*request.Description, constants.MsgLen)
	}
	patch := client.MergeFrom(version)
	data, err := patch.Data(versionCopy)
	if err != nil {
		klog.Error("create patch failed", err)
		return err
	}

	//data == "{}", need not to patch
	if len(data) == 2 {
		return nil
	}

	//_, err = c.appClient.HelmApplicationVersions().Patch(context.TODO(), id, patch.Type(), data, metav1.PatchOptions{})
	_, err = c.appVersionClient.Patch(context.TODO(), id, patch.Type(), data, metav1.PatchOptions{})

	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *applicationOperator) ListAppVersions(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	versions, err := c.listAppVersions(conditions.Match[AppId])

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	//if conditions.Match[Keyword] != "" {
	//	versions = helmApplicationFilter(conditions.Match[Keyword], repos)
	//}

	status := strings2.Split(conditions.Match[Status], "|")
	versions = filterAppVersionByState(versions, status)
	if reverse {
		sort.Sort(sort.Reverse(AppVersions(versions)))
	} else {
		sort.Sort(AppVersions(versions))
	}

	items := make([]interface{}, 0, int(math.Min(float64(limit), float64(len(versions)))))

	for i, j := offset, 0; i < len(versions) && j < limit; {
		items = append(items, convertAppVersion(versions[i]))
		i++
		j++
	}
	return &models.PageableResponse{Items: items, TotalCount: len(versions)}, nil
}

func (c *applicationOperator) ListAppVersionReviews(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	var allStatus []string
	if status := conditions.Match[Status]; status != "" {
		allStatus = strings2.Split(status, "|")
	}

	appVersions, err := c.versionLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	filtered := make([]*v1beta1.FederatedHelmApplicationVersion, 0, len(appVersions)/2)
	for _, version := range appVersions {
		filledVersion, err := c.fillAppVersionAudit(version)
		if err != nil {
			klog.Errorf("get app version %s audit failed, error: %s", version.Name, err)
			return nil, err
		}
		if sliceutil.HasString(allStatus, filledVersion.Spec.Template.AuditSpec.State) {
			filtered = append(filtered, filledVersion)
		}
	}

	items := make([]interface{}, 0)

	for i, j := offset, 0; i < len(filtered) && j < limit; {
		//TODO, appversion status
		review := convertAppVersionReview(filtered[i])
		items = append(items, review)
		i++
		j++
	}

	return &models.PageableResponse{Items: items, TotalCount: len(filtered)}, nil
}

func (c *applicationOperator) ListAppVersionAudits(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	appId := conditions.Match[AppId]
	versionId := conditions.Match[VersionId]

	var versions []*v1beta1.FederatedHelmApplicationVersion
	var err error
	if versionId == "" {
		ls := map[string]string{
			constants.ChartApplicationIdLabelKey: appId,
		}
		versions, err = c.versionLister.List(helpers.MapAsLabelSelector(ls))
		if err != nil {
			klog.Errorf("get app %s failed, error: %s", appId, err)
		}
	} else {
		version, err := c.versionLister.Get(versionId)
		if err != nil {
			klog.Errorf("get app version %s failed, error: %s", versionId, err)
		}
		versions = []*v1beta1.FederatedHelmApplicationVersion{version}
	}

	var allAudits []*AppVersionAudit
	for _, item := range versions {
		a, err := c.auditLister.Get(item.Name)
		if err == nil {
			item.Spec.Template.AuditSpec = a.Spec
		}
		audits := convertAppVersionAudit(item)
		allAudits = append(allAudits, audits...)
	}

	sort.Sort(AppVersionAuditList(allAudits))

	items := make([]interface{}, 0, limit)

	for i, j := offset, 0; i < len(allAudits) && j < limit; {
		items = append(items, allAudits[i])
		i++
		j++
	}

	return &models.PageableResponse{Items: items, TotalCount: len(allAudits)}, nil
}

func (c *applicationOperator) fillAppVersionAudit(appVersion *v1beta1.FederatedHelmApplicationVersion) (*v1beta1.FederatedHelmApplicationVersion, error) {
	audit, err := c.auditLister.Get(appVersion.Name)
	if err != nil {
		return nil, err
	}
	appVersion.Spec.Template.AuditSpec = audit.Spec
	return appVersion, nil
}

func (c *applicationOperator) DoAppVersionAction(versionId string, request *ActionRequest) error {
	var err error
	t := metav1.Now()
	var audit = v1alpha1.Audit{
		Message:  request.Message,
		Operator: request.Username,
		Time:     t,
	}
	state := constants.StateDraft

	version, err := c.versionLister.Get(versionId)
	if err != nil {
		klog.Errorf("get app version %s failed, error: %s", versionId, err)
		return err
	}

	helmAudit, err := c.auditLister.Get(version.Name)
	if err != nil {
		klog.Errorf("get app version audit %s failed, error: %s", versionId, err)
		return err
	}

	//rls, err := c.rlsLister.List(helpers.MapAsLabelSelector(map[string]string{constants.ChartApplicationVersionIdLabelKey: versionId}))
	//if err != nil && apiErrors.IsNotFound(err) {
	//	return err
	//}
	//
	//if len(rls) > 0 && request.Action == ActionSuspend {
	//	return errors.New("helm application has release")
	//}

	switch request.Action {
	case ActionCancel:
		if helmAudit.Spec.State != constants.StateSubmitted {
		}
		state = constants.StateDraft
		audit.State = constants.StateDraft
	case ActionPass:
		if helmAudit.Spec.State != constants.StateSubmitted {

		}
		state = constants.StatePassed
		audit.State = constants.StatePassed
	case ActionRecover:
		if helmAudit.Spec.State != constants.StateSuspended {

		}
		state = constants.StateActive
		audit.State = constants.StateActive
	case ActionReject:
		if helmAudit.Spec.State != constants.StateSubmitted {

		}
		state = constants.StateRejected
		audit.State = constants.StateRejected
	case ActionSubmit:
		if helmAudit.Spec.State != constants.StateDraft {

		}
		state = constants.StateSubmitted
		audit.State = constants.StateSubmitted
	case ActionSuspend:
		if helmAudit.Spec.State != constants.StateActive {

		}
		state = constants.StateSuspended
		audit.State = constants.StateSuspended
	case ActionRelease:
		//release to app store
		if helmAudit.Spec.State != constants.StatePassed {
			//err =
		}
		state = constants.StateActive
		audit.State = constants.StateActive
	default:
		err = errors.New("action not support")
	}

	_ = state
	if err != nil {
		klog.Error(err)
		return err
	}

	//version, err = c.updateVersionStatus(version, state, &audit)
	err = c.updateAppVersionAudit(helmAudit, &audit)

	if err != nil {
		klog.Errorf("update app version audit [%s] failed, error: %s", versionId, err)
		return err
	}

	if request.Action == ActionRelease || request.Action == ActionRecover {
		//if we release a new helm application version, we need update the spec in helm application copy
		app, err := c.appLister.Get(version.GetHelmApplicationId())
		if err != nil {
			return err
		}
		appInStore, err := c.appLister.Get(fmt.Sprintf("%s%s", version.GetHelmApplicationId(), constants.HelmApplicationAppStoreSuffix))
		if err != nil {
			if apiErrors.IsNotFound(err) {
				//controller-manager will create application in app store
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
func (c *applicationOperator) updateAppVersionAudit(helmAudit *v1alpha1.HelmAudit, auditInfo *v1alpha1.Audit) error {
	if len(helmAudit.Spec.Audit) > 0 {
		helmAudit.Spec.Audit = append([]v1alpha1.Audit{*auditInfo}, helmAudit.Spec.Audit...)
	} else {
		helmAudit.Spec.Audit = []v1alpha1.Audit{*auditInfo}
	}
	if len(helmAudit.Spec.Audit) > constants.HelmAuditLen {
		helmAudit.Spec.Audit = helmAudit.Spec.Audit[:constants.HelmAuditLen:constants.HelmAuditLen]
	}
	helmAudit.Spec.State = auditInfo.State

	_, err := c.auditClient.Update(context.TODO(), helmAudit, metav1.UpdateOptions{})
	return err
}

//Create helmApplicationVersion and helmAudit
func (c *applicationOperator) createApplicationVersionWithAudit(ver *v1beta1.FederatedHelmApplicationVersion) (*v1beta1.FederatedHelmApplicationVersion, error) {
	ls := map[string]string{
		//constants.WorkspaceLabelKey:          ver.GetWorkspace(),
		constants.ChartApplicationIdLabelKey: ver.GetHelmApplicationId(),
	}

	list, err := c.versionLister.List(helpers.MapAsLabelSelector(ls))

	if err != nil && !apiErrors.IsNotFound(err) {
		return nil, err
	}

	if len(list) > 0 {
		verName := ver.GetVersionName()
		for _, v := range list {
			if verName == v.GetVersionName() {
				klog.V(2).Infof("helm application version: %s exist", verName)
				return nil, ItemExists
			}
		}
	}

	version, err := c.appVersionClient.Create(context.TODO(), ver, metav1.CreateOptions{})
	if err == nil {
		klog.V(4).Infof("create helm application %s version success", version.Name)
		//crate application version audit
		audit := buildApplicationVersionAudit(version)
		audit, err := c.auditClient.Create(context.TODO(), audit, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("create helm audit for %s failed, error: %s", version.Name, err)
		} else {
			klog.V(4).Infof("create helm audit for %s success", version.Name)
		}
	}

	return version, err
}

//func (c *applicationOperator) updateVersionStatus(version *v1alpha1.HelmApplicationVersion, state string, status *v1alpha1.Audit) (*v1alpha1.HelmApplicationVersion, error) {
//	version.Status.State = state
//
//	states := append([]v1alpha1.Audit{*status}, version.Status.Audit...)
//	if len(version.Status.Audit) >= constants.HelmRepoSyncStateLen {
//		//strip the last item
//		states = states[:constants.HelmRepoSyncStateLen:constants.HelmRepoSyncStateLen]
//	}
//
//	version.Status.Audit = states
//	//version, err := c.appClient.HelmApplicationVersions().UpdateStatus(context.TODO(), version, metav1.UpdateOptions{})
//	version, err := c.appClient.UpdateStatus(version)
//
//	return version, err
//}

func (c *applicationOperator) GetAppVersionFiles(versionId string, request *GetAppVersionFilesRequest) (*GetAppVersionPackageFilesResponse, error) {
	c.cachedRepos.RLock()
	defer c.cachedRepos.RUnlock()
	var reader io.Reader
	var version *v1beta1.FederatedHelmApplicationVersion
	var err error

	version, err = c.getHelmAppVersion(versionId)
	if err != nil {
		return nil, err
	}
	template := version.Spec.Template
	if len(template.Spec.Data) == 0 {
		repo := c.getHelmRepo(version.GetHelmRepoId())
		if len(template.Spec.URLs) == 0 {
			return nil, nil
		}
		url := template.Spec.URLs[0]
		if !(strings2.HasPrefix(url, "https://") || strings2.HasPrefix(url, "http://")) {
			url = repo.Spec.Url + "/" + url
		}
		buf, err := helmrepoindex.LoadChart(context.TODO(), url, &repo.Spec.Credential)
		if err != nil {
			klog.Errorf("load chart version [%s] failed,  error : %s", versionId, err)
			return nil, err
		}
		reader = buf
	} else {
		//data, err = base64.StdEncoding.DecodeString(string(version.Spec.Data))
		reader = bytes.NewBuffer([]byte(template.Spec.Data))
	}

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		klog.Errorf("read chart version [%s] failed, error: %s", versionId, err)
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
			name := strings2.TrimLeft(header.Name, fmt.Sprintf("%s/", version.GetTrueName()))
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

func (c *applicationOperator) getHelmRepo(repoId string) *v1alpha1.HelmRepo {
	c.cachedRepos.RLock()
	defer c.cachedRepos.RUnlock()
	if repo, exists := c.cachedRepos.repos[repoId]; exists {
		return repo
	}

	return nil
}

func (c *applicationOperator) getHelmAppVersion(versionId string) (*v1beta1.FederatedHelmApplicationVersion, error) {
	c.cachedRepos.RLock()
	if version, exists := c.cachedRepos.versions[versionId]; exists {
		c.cachedRepos.RUnlock()
		return version, nil
	}

	c.cachedRepos.RUnlock()
	version, err := c.versionLister.Get(versionId)
	if err != nil {
		return nil, err
	}

	//data, err = base64.StdEncoding.DecodeString(string(version.Spec.Data))
	return version, nil
}

func (c *applicationOperator) listAppVersions(appId string) (ret []*v1beta1.FederatedHelmApplicationVersion, err error) {
	c.cachedRepos.RLock()

	//list app version from cache
	if _, exists := c.cachedRepos.apps[appId]; exists {
		ret = c.cachedRepos.listAppVersions(appId)
		c.cachedRepos.RUnlock()
		return
	}

	c.cachedRepos.RUnlock()
	ret, err = c.versionLister.List(helpers.MapAsLabelSelector(map[string]string{constants.ChartApplicationIdLabelKey: appId}))
	if err != nil && !apiErrors.IsNotFound(err) {
		klog.Error(err)
		return nil, err
	}

	for _, version := range ret {
		audit, err := c.auditLister.Get(version.Name)
		if err == nil {
			version.Spec.Template.AuditSpec = audit.Spec
		}
	}

	return
}
