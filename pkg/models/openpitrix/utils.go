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
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"path"
	"regexp"
	"strings"
	"time"
)

func convertRepoEvent(meta *metav1.ObjectMeta, state *v1alpha1.HelmRepoSyncState) *RepoEvent {
	if meta == nil || state == nil {
		return nil
	}
	out := RepoEvent{}
	date := strfmt.DateTime(time.Unix(state.SyncTime.Unix(), 0))
	out.CreateTime = &date
	out.RepoId = meta.Name
	out.RepoEventId = ""
	out.Result = state.Message
	out.Status = state.State
	out.StatusTime = out.CreateTime

	return &out
}

func convertAppVersionAudit(appVersion *v1alpha1.HelmApplicationVersion) []*AppVersionAudit {
	if appVersion == nil {
		return nil
	}
	var audits []*AppVersionAudit
	for _, helmAudit := range appVersion.Status.Audit {
		var audit AppVersionAudit
		audit.AppId = appVersion.GetHelmApplicationId()
		audit.Operator = helmAudit.Operator
		audit.Message = helmAudit.Message
		audit.Status = helmAudit.State
		date := strfmt.DateTime(time.Unix(helmAudit.Time.Unix(), 0))
		audit.StatusTime = &date
		audit.VersionId = appVersion.Name
		audit.VersionType = "helm"
		audit.VersionName = appVersion.GetVersionName()
		audit.Operator = helmAudit.Operator
		audit.OperatorType = helmAudit.OperatorType

		audits = append(audits, &audit)
	}

	return audits
}

type HelmReleaseList []*v1alpha1.HelmRelease

func (l HelmReleaseList) Len() int      { return len(l) }
func (l HelmReleaseList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l HelmReleaseList) Less(i, j int) bool {
	var t1, t2 time.Time
	if l[i].Status.LastUpdate.IsZero() {
		t1 = l[i].CreationTimestamp.Time
	} else {
		t1 = l[i].Status.LastUpdate.Time
	}

	if l[j].Status.LastUpdate.IsZero() {
		t2 = l[j].CreationTimestamp.Time
	} else {
		t2 = l[j].Status.LastUpdate.Time
	}

	if t1.After(t2) {
		return true
	} else if t1.Before(t2) {
		return false
	} else {
		return l[i].Name > l[j].Name
	}
}

type AppVersionAuditList []*AppVersionAudit

func (l AppVersionAuditList) Len() int      { return len(l) }
func (l AppVersionAuditList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l AppVersionAuditList) Less(i, j int) bool {
	t1 := l[i].StatusTime.String()
	t2 := l[j].StatusTime.String()
	if t1 > t2 {
		return true
	} else if t1 < t2 {
		return false
	} else {
		n1 := l[i].VersionName
		n2 := l[j].VersionName
		return n1 < n2
	}
}

// copy from openpitrix
func matchPackageFailedError(err error, res *ValidatePackageResponse) {
	var errStr = err.Error()
	var matchedError = ""
	var errorDetails = make(map[string]string)
	switch {
	// Helm errors
	case strings.HasPrefix(errStr, "no files in chart archive"),
		strings.HasPrefix(errStr, "no files in app archive"):

		matchedError = "no files in package"

	case strings.HasPrefix(errStr, "chart yaml not in base directory"),
		strings.HasPrefix(errStr, "chart metadata (Chart.yaml) missing"):

		errorDetails["Chart.yaml"] = "not found"

	case strings.HasPrefix(errStr, "invalid chart (Chart.yaml): name must not be empty"):

		errorDetails["Chart.yaml"] = "package name must not be empty"

	case strings.HasPrefix(errStr, "values.toml is illegal"):

		errorDetails["values.toml"] = errStr

	case strings.HasPrefix(errStr, "error reading"):

		matched := regexp.MustCompile("error reading (.+): (.+)").FindStringSubmatch(errStr)
		if len(matched) > 0 {
			errorDetails[matched[1]] = matched[2]
		}

		// Devkit errors
	case strings.HasPrefix(errStr, "[package.json] not in base directory"):

		errorDetails["package.json"] = "not found"

	case strings.HasPrefix(errStr, "missing file ["):

		matched := regexp.MustCompile("missing file \\[(.+)]").FindStringSubmatch(errStr)
		if len(matched) > 0 {
			errorDetails[matched[1]] = "not found"
		}

	case strings.HasPrefix(errStr, "failed to parse"),
		strings.HasPrefix(errStr, "failed to render"),
		strings.HasPrefix(errStr, "failed to load"),
		strings.HasPrefix(errStr, "failed to decode"):

		matched := regexp.MustCompile("failed to (.+) (.+): (.+)").FindStringSubmatch(errStr)
		if len(matched) > 0 {
			errorDetails[matched[2]] = fmt.Sprintf("%s failed, %s", matched[1], matched[3])
		}

	default:
		matchedError = errStr
	}
	if len(errorDetails) > 0 {
		res.ErrorDetails = errorDetails
	}
	if len(matchedError) > 0 {
		res.Error = matchedError
	}
}

func convertCategory(in *v1alpha1.HelmCategory) *Category {
	if in == nil {
		return nil
	}
	out := &Category{}
	out.Description = in.Spec.Description
	out.Name = in.Spec.Name
	out.CategoryID = in.Name
	t := strfmt.DateTime(in.CreationTimestamp.Time)
	out.CreateTime = &t
	if in.Spec.Locale == "" {
		out.Locale = "{}"
	} else {
		out.Locale = in.Spec.Locale
	}
	total := in.Status.Total
	out.AppTotal = &total

	return out
}

func convertApplication(rls *v1alpha1.HelmRelease, rlsInfos []*resource.Info) *Application {
	app := &Application{}
	app.Name = rls.Spec.ChartName
	cluster := &Cluster{}
	cluster.ClusterId = rls.Name
	cluster.Owner = rls.GetCreator()
	cluster.Zone = rls.GetRlsNamespace()
	cluster.Status = rls.Status.State
	cluster.Env = string(rls.Spec.Values)
	if cluster.Status == "" {
		cluster.Status = v1alpha1.HelmStatusPending
	}
	cluster.AdditionalInfo = rls.Status.Message
	cluster.Description = rls.Spec.Description
	dt := strfmt.DateTime(rls.CreationTimestamp.Time)
	cluster.CreateTime = &dt
	if !rls.Status.LastUpdate.Time.IsZero() {
		ut := strfmt.DateTime(rls.Status.LastUpdate.Time)
		cluster.StatusTime = &ut
	} else {
		cluster.StatusTime = &dt
	}
	cluster.AppId = rls.Spec.ApplicationId
	cluster.VersionId = rls.Spec.ApplicationVersionId
	cluster.Name = rls.GetTrueName()

	if rls.GetRlsCluster() != "" {
		cluster.RuntimeId = rls.GetRlsCluster()
	} else {
		cluster.RuntimeId = "default"
	}

	app.Cluster = cluster
	app.Version = &AppVersion{
		AppId:     rls.Spec.ApplicationId,
		VersionId: rls.Spec.ApplicationVersionId,
		Name:      rls.GetChartVersionName(),
	}
	app.App = &App{
		AppId:     rls.Spec.ApplicationId,
		ChartName: rls.Spec.ChartName,
		Name:      rls.Spec.ChartName,
	}

	app.ReleaseInfo = make([]runtime.Object, 0, len(rlsInfos))
	if rlsInfos != nil {
		for _, info := range rlsInfos {
			app.ReleaseInfo = append(app.ReleaseInfo, info.Object)
		}
	}

	return app
}

func convertApp(app *v1alpha1.HelmApplication, versions []*v1alpha1.HelmApplicationVersion, ctg *v1alpha1.HelmCategory, rlsCount int) *App {
	if app == nil {
		return nil
	}
	out := &App{}

	out.AppId = app.Name
	out.Name = app.GetTrueName()

	date := strfmt.DateTime(app.CreationTimestamp.Time)
	out.CreateTime = &date
	if app.Status.StatusTime != nil {
		s := strfmt.DateTime(app.Status.StatusTime.Time)
		out.StatusTime = &s
	} else {
		out.StatusTime = out.CreateTime
	}

	if app.Status.UpdateTime == nil {
		out.UpdateTime = out.CreateTime
	} else {
		u := strfmt.DateTime(app.Status.UpdateTime.Time)
		out.UpdateTime = &u
	}

	out.Status = app.Status.State
	if out.Status == "" {
		out.Status = v1alpha1.StateDraft
	}
	out.Abstraction = app.Spec.Abstraction
	out.Description = app.Spec.Description

	if len(app.Spec.Attachments) > 0 {
		out.Screenshots = strings.Join(app.Spec.Attachments, ",")
	}
	out.Home = app.Spec.AppHome
	out.Icon = app.Spec.Icon

	if ctg != nil {
		ct := strfmt.DateTime(ctg.CreationTimestamp.Time)
		rc := ResourceCategory{
			CategoryId: ctg.Name,
			Name:       ctg.GetTrueName(),
			CreateTime: &ct,
			Locale:     ctg.Spec.Locale,
		}
		if ctg.Spec.Locale == "" {
			rc.Locale = "{}"
		} else {
			rc.Locale = ctg.Spec.Locale
		}
		rc.Status = "enabled"

		out.CategorySet = AppCategorySet{&rc}
	} else {
		out.CategorySet = AppCategorySet{}
	}

	for _, version := range versions {
		if app.Status.LatestVersion == version.GetVersionName() {
			// find the latest version, and convert its format
			out.LatestAppVersion = convertAppVersion(version)
			break
		}
	}

	if out.LatestAppVersion == nil {
		out.LatestAppVersion = &AppVersion{}
	}

	out.AppVersionTypes = "helm"
	out.Isv = app.GetWorkspace()

	out.ClusterTotal = &rlsCount
	out.Owner = app.GetCreator()

	return out
}

func filterAppVersionByState(versions []*v1alpha1.HelmApplicationVersion, states []string) []*v1alpha1.HelmApplicationVersion {
	if len(states) == 0 {
		return versions
	}

	var j = 0
	for i := 0; i < len(versions); i++ {
		state := versions[i].State()
		if sliceutil.HasString(states, state) {
			if i != j {
				versions[j] = versions[i]
			}
			j++
		}
	}

	versions = versions[:j:j]
	return versions
}

func convertAppVersion(in *v1alpha1.HelmApplicationVersion) *AppVersion {
	if in == nil {
		return nil
	}
	out := AppVersion{}
	out.AppId = in.GetHelmApplicationId()
	out.Active = true
	t := in.CreationTimestamp.Time
	date := strfmt.DateTime(t)
	out.CreateTime = &date
	if len(in.Status.Audit) > 0 {
		t = in.Status.Audit[0].Time.Time
		updateDate := strfmt.DateTime(t)
		out.UpdateTime = &updateDate
	} else {
		out.UpdateTime = &date
	}
	if in.Spec.Metadata != nil {
		out.Description = in.Spec.Description
		out.Icon = in.Spec.Icon
	}

	out.Status = in.State()
	out.Owner = in.GetCreator()
	out.Name = in.GetVersionName()
	out.VersionId = in.GetHelmApplicationVersionId()
	return &out
}

func convertRepo(in *v1alpha1.HelmRepo) *Repo {
	if in == nil {
		return nil
	}
	out := Repo{}

	out.RepoId = in.GetHelmRepoId()
	out.Name = in.GetTrueName()

	out.Status = in.Status.State
	date := strfmt.DateTime(time.Unix(in.CreationTimestamp.Unix(), 0))
	out.CreateTime = &date

	out.Description = in.Spec.Description

	out.URL = in.Spec.Url
	return &out
}

type HelmCategoryList []*v1alpha1.HelmCategory

func (l HelmCategoryList) Len() int      { return len(l) }
func (l HelmCategoryList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l HelmCategoryList) Less(i, j int) bool {
	t1 := l[i].CreationTimestamp.UnixNano()
	t2 := l[j].CreationTimestamp.UnixNano()
	if t1 > t2 {
		return true
	} else if t1 < t2 {
		return false
	} else {
		n1 := l[i].Spec.Name
		n2 := l[j].Spec.Name
		return n1 < n2
	}
}

type HelmApplicationList []*v1alpha1.HelmApplication

func (l HelmApplicationList) Len() int      { return len(l) }
func (l HelmApplicationList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l HelmApplicationList) Less(i, j int) bool {
	t1 := l[i].CreationTimestamp.UnixNano()
	t2 := l[j].CreationTimestamp.UnixNano()
	if t1 < t2 {
		return true
	} else if t1 > t2 {
		return false
	} else {
		n1 := l[i].GetTrueName()
		n2 := l[j].GetTrueName()
		return n1 < n2
	}
}

type AppVersionReviews []*v1alpha1.HelmApplicationVersion

// Len returns the length.
func (c AppVersionReviews) Len() int { return len(c) }

// Swap swaps the position of two items in the versions slice.
func (c AppVersionReviews) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (c AppVersionReviews) Less(a, b int) bool {
	aVersion := c[a]
	bVersion := c[b]

	if len(aVersion.Status.Audit) > 0 && len(bVersion.Status.Audit) > 0 {
		t1 := aVersion.Status.Audit[0].Time
		t2 := bVersion.Status.Audit[0].Time
		if t1.Before(&t2) {
			return true
		} else if t2.Before(&t1) {
			return false
		}
	}

	i, err := semver.NewVersion(aVersion.GetSemver())
	if err != nil {
		return true
	}
	j, err := semver.NewVersion(bVersion.GetSemver())
	if err != nil {
		return false
	}
	if i.Equal(j) {
		return aVersion.CreationTimestamp.Before(&bVersion.CreationTimestamp)
	}
	return j.LessThan(i)
}

type AppVersions []*v1alpha1.HelmApplicationVersion

// Len returns the length.
func (c AppVersions) Len() int { return len(c) }

// Swap swaps the position of two items in the versions slice.
func (c AppVersions) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (c AppVersions) Less(a, b int) bool {
	// Failed parse pushes to the back.
	aVersion := c[a]
	bVersion := c[b]
	i, err := semver.NewVersion(aVersion.GetSemver())
	if err != nil {
		return true
	}
	j, err := semver.NewVersion(bVersion.GetSemver())
	if err != nil {
		return false
	}
	if i.Equal(j) {
		return aVersion.CreationTimestamp.Before(&bVersion.CreationTimestamp)
	}
	return i.LessThan(j)
}

// buildApplicationVersion  build an application version
// packageData base64 encoded package data
func buildApplicationVersion(app *v1alpha1.HelmApplication, chrt helmrepoindex.VersionInterface, chartPackage *string, creator string) *v1alpha1.HelmApplicationVersion {
	// create helm application version resource
	t := metav1.Now()
	ver := &v1alpha1.HelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: creator,
			},
			Name: idutils.GetUuid36(v1alpha1.HelmApplicationVersionIdPrefix),
			Labels: map[string]string{
				constants.ChartApplicationIdLabelKey: app.GetHelmApplicationId(),
				constants.WorkspaceLabelKey:          app.GetWorkspace(),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:        app.GetUID(),
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       "HelmApplication",
					Name:       app.Name,
				},
			},
		},
		Spec: v1alpha1.HelmApplicationVersionSpec{
			Metadata: &v1alpha1.Metadata{
				Version:     chrt.GetVersion(),
				AppVersion:  chrt.GetAppVersion(),
				Name:        chrt.GetName(),
				Icon:        chrt.GetIcon(),
				Home:        chrt.GetHome(),
				Description: stringutils.ShortenString(chrt.GetDescription(), v1alpha1.MsgLen),
			},
			Created: &t,
			// set data to nil before save app version to etcd
			Data: []byte(*chartPackage),
		},
		Status: v1alpha1.HelmApplicationVersionStatus{
			State: v1alpha1.StateDraft,
			Audit: []v1alpha1.Audit{
				{
					State:    v1alpha1.StateDraft,
					Time:     t,
					Operator: creator,
				},
			},
		},
	}

	return ver
}

func filterAppByName(app *v1alpha1.HelmApplication, namePart string) bool {
	if len(namePart) == 0 {
		return true
	}

	name := app.GetTrueName()
	if strings.Contains(name, namePart) {
		return true
	}
	return false
}

func filterAppByStates(app *v1alpha1.HelmApplication, state []string) bool {
	if len(state) == 0 {
		return true
	}
	st := app.Status.State
	// default value is draft
	if st == "" {
		st = v1alpha1.StateDraft
	}
	if sliceutil.HasString(state, st) {
		return true
	}
	return false
}

func filterAppReviews(versions []*v1alpha1.HelmApplicationVersion, conditions *params.Conditions) []*v1alpha1.HelmApplicationVersion {
	if conditions == nil || len(conditions.Match) == 0 || len(versions) == 0 {
		return versions
	}

	curr := 0
	for i := 0; i < len(versions); i++ {
		if conditions.Match[Keyword] != "" {
			if !(strings.Contains(versions[i].Spec.Name, conditions.Match[Keyword])) {
				continue
			}
		}

		if conditions.Match[Status] != "" {
			states := strings.Split(conditions.Match[Status], "|")
			state := versions[i].State()
			if !sliceutil.HasString(states, state) {
				continue
			}
		}

		if curr != i {
			versions[curr] = versions[i]
		}
		curr++
	}

	return versions[:curr:curr]
}

func filterAppVersions(versions []*v1alpha1.HelmApplicationVersion, conditions *params.Conditions) []*v1alpha1.HelmApplicationVersion {
	if conditions == nil || len(conditions.Match) == 0 || len(versions) == 0 {
		return versions
	}

	curr := 0
	for i := 0; i < len(versions); i++ {
		if conditions.Match[Keyword] != "" {
			if !(strings.Contains(versions[i].Spec.Version, conditions.Match[Keyword]) ||
				strings.Contains(versions[i].Spec.AppVersion, conditions.Match[Keyword])) {
				continue
			}
		}

		if conditions.Match[Status] != "" {
			states := strings.Split(conditions.Match[Status], "|")
			state := versions[i].State()
			if !sliceutil.HasString(states, state) {
				continue
			}
		}

		if curr != i {
			versions[curr] = versions[i]
		}
		curr++
	}

	return versions[:curr:curr]
}

func filterApps(apps []*v1alpha1.HelmApplication, conditions *params.Conditions) []*v1alpha1.HelmApplication {
	if conditions == nil || len(conditions.Match) == 0 || len(apps) == 0 {
		return apps
	}

	curr := 0
	for i := 0; i < len(apps); i++ {
		if conditions.Match[Keyword] != "" {
			fv := filterAppByName(apps[i], conditions.Match[Keyword])
			if !fv {
				continue
			}
		}

		if conditions.Match[Status] != "" {
			states := strings.Split(conditions.Match[Status], "|")
			fv := filterAppByStates(apps[i], states)
			if !fv {
				continue
			}
		}
		if curr != i {
			apps[curr] = apps[i]
		}
		curr++
	}

	return apps[:curr:curr]
}

func filterReleaseByStates(rls *v1alpha1.HelmRelease, state []string) bool {
	if len(state) == 0 {
		return true
	}
	st := rls.Status.State
	if st == "" {
		st = v1alpha1.HelmStatusCreating
	}
	if sliceutil.HasString(state, st) {
		return true
	}
	return false
}

func filterReleasesWithAppVersions(releases []*v1alpha1.HelmRelease, appVersions map[string]*v1alpha1.HelmApplicationVersion) []*v1alpha1.HelmRelease {
	if appVersions == nil || len(appVersions) == 0 || len(releases) == 0 {
		return []*v1alpha1.HelmRelease{}
	}

	curr := 0
	for i := 0; i < len(releases); i++ {
		if _, exists := appVersions[releases[i].Spec.ApplicationVersionId]; exists {
			if curr != i {
				releases[curr] = releases[i]
			}
			curr++
		}
	}

	return releases[:curr:curr]
}

func filterReleases(releases []*v1alpha1.HelmRelease, conditions *params.Conditions) []*v1alpha1.HelmRelease {
	if conditions == nil || len(conditions.Match) == 0 || len(releases) == 0 {
		return releases
	}

	curr := 0
	for i := 0; i < len(releases); i++ {
		keyword := conditions.Match[Keyword]
		if keyword != "" {
			fv := strings.Contains(releases[i].GetTrueName(), keyword) ||
				strings.Contains(releases[i].Spec.ChartVersion, keyword) ||
				strings.Contains(releases[i].Spec.ChartAppVersion, keyword)
			if !fv {
				continue
			}
		}

		if conditions.Match[Status] != "" {
			states := strings.Split(conditions.Match[Status], "|")
			fv := filterReleaseByStates(releases[i], states)
			if !fv {
				continue
			}
		}
		if curr != i {
			releases[curr] = releases[i]
		}
		curr++
	}

	return releases[:curr:curr]
}

func dataKeyInStorage(workspace, id string) string {
	return path.Join(workspace, id)
}

func attachmentKeyInStorage(ws, id string) string {
	return path.Join(ws, id)
}

func convertAppVersionReview(appVersion *v1alpha1.HelmApplicationVersion) *AppVersionReview {
	review := &AppVersionReview{}
	status := appVersion.Status
	review.Reviewer = status.Audit[0].Operator
	review.ReviewId = status.Audit[0].Operator
	review.Status = appVersion.Status.State
	review.AppId = appVersion.GetHelmApplicationId()
	review.VersionID = appVersion.GetHelmApplicationVersionId()
	review.Phase = AppVersionReviewPhaseOAIGen{}
	review.VersionName = appVersion.GetVersionName()

	review.StatusTime = strfmt.DateTime(status.Audit[0].Time.Time)
	review.AppName = appVersion.GetTrueName()
	return review
}

func parseChartVersionName(name string) (version, appVersion string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}

	parts := strings.Split(name, "[")
	if len(parts) == 1 {
		return parts[0], ""
	}

	version = strings.TrimSpace(parts[0])

	appVersion = strings.Trim(parts[1], "]")
	appVersion = strings.TrimSpace(appVersion)
	return
}
