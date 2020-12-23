package openpitrix

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	strings2 "k8s.io/utils/strings"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"regexp"
	"sort"
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

func toFederateHelmApplication(app *v1alpha1.HelmApplication) *v1beta1.FederatedHelmApplication {
	fed := &v1beta1.FederatedHelmApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:   app.Name,
			Labels: app.Labels,
		},
		Spec: v1beta1.FederatedHelmApplicationSpec{
			Template: v1beta1.HelmApplicationTemplate{
				ObjectMeta: app.ObjectMeta,
				Spec:       app.Spec,
			},
			Placement: v1beta1.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{},
				},
			},
		},
	}
	return fed
}

func toFederateHelmApplicationVersion(appVersion *v1alpha1.HelmApplicationVersion) *v1beta1.FederatedHelmApplicationVersion {
	fed := &v1beta1.FederatedHelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: appVersion.Name,
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: appVersion.GetCreator(),
			},
			Labels: appVersion.Labels,
		},
		Spec: v1beta1.FederatedHelmApplicationVersionSpec{
			Template: v1beta1.HelmApplicationVersionTemplate{
				ObjectMeta: appVersion.ObjectMeta,
				Spec:       appVersion.Spec,
			},
			Placement: v1beta1.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{},
				},
			},
		},
	}
	return fed
}

func convertAppVersionAudit(in *v1beta1.FederatedHelmApplicationVersion) []*AppVersionAudit {
	if in == nil {
		return nil
	}
	var audits []*AppVersionAudit
	template := in.Spec.Template
	for ind := range template.AuditSpec.Audit {
		var audit AppVersionAudit
		a := &template.AuditSpec.Audit[ind]
		audit.AppId = in.GetHelmApplicationId()
		audit.Operator = a.Operator
		audit.Message = a.Message
		audit.Status = a.State
		date := strfmt.DateTime(time.Unix(a.Time.Unix(), 0))
		audit.StatusTime = &date
		audit.VersionId = in.Name
		audit.VersionType = "helm"
		audit.VersionName = in.GetVersionName()
		audit.Operator = a.Operator
		audit.OperatorType = a.OperatorType

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
		return l[i].Name < l[j].Name
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

//copy from openpitrix
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

func convertApplication(clusterName string, rls *v1alpha1.HelmRelease, version *v1beta1.FederatedHelmApplicationVersion, rlsInfos []*resource.Info) *Application {
	app := &Application{}
	app.Name = rls.Spec.Name
	cluster := &Cluster{}
	cluster.Zone = rls.Namespace
	cluster.ClusterId = rls.Name
	cluster.Owner = rls.GetCreator()
	cluster.Status = rls.Status.State
	cluster.Env = string(rls.Spec.Values)
	//cluster.
	if cluster.Status == "" {
		cluster.Status = constants.HelmStatusPending
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
	cluster.Name = rls.Name

	if clusterName != "" {
		cluster.RuntimeId = clusterName
	} else {
		cluster.RuntimeId = "default"
	}
	//cluster.

	app.Cluster = cluster
	app.Version = convertAppVersion(version)
	app.App = &App{
		AppId:     version.GetVersionName(),
		ChartName: version.GetTrueName(),
		Name:      version.GetTrueName(),
	}

	app.ReleaseInfo = make([]runtime.Object, 0, len(rlsInfos))
	if rlsInfos != nil {
		for _, info := range rlsInfos {
			app.ReleaseInfo = append(app.ReleaseInfo, info.Object)
		}
	}

	return app
}

func convertApp(in *v1beta1.FederatedHelmApplication, versions []*v1beta1.FederatedHelmApplicationVersion, ctg *v1alpha1.HelmCategory, rlsCount int) *App {
	if in == nil {
		return nil
	}
	out := &App{}

	out.AppId = in.Name
	out.Name = in.GetTrueName()

	template := in.Spec.Template
	date := strfmt.DateTime(in.CreationTimestamp.Time)
	out.CreateTime = &date
	if template.Spec.StatusTime != nil {
		s := strfmt.DateTime(template.Spec.StatusTime.Time)
		out.StatusTime = &s
	} else {
		out.StatusTime = out.CreateTime
	}

	if template.Spec.UpdateTime == nil {
		out.UpdateTime = out.CreateTime
	} else {
		u := strfmt.DateTime(template.Spec.UpdateTime.Time)
		out.UpdateTime = &u
	}

	out.Status = template.Spec.Status
	if out.Status == "" {
		out.Status = constants.StateDraft
	}
	out.Abstraction = template.Spec.Abstraction
	out.Description = template.Spec.Description

	if len(template.Spec.Attachments) > 0 {
		out.Screenshots = strings.Join(template.Spec.Attachments, ",")
	}
	out.Home = template.Spec.AppHome
	out.Icon = template.Spec.Icon

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
	if versions != nil && len(versions) > 0 {
		sort.Sort(AppVersions(versions))
		out.LatestAppVersion = convertAppVersion(versions[0])
	} else {
		out.LatestAppVersion = &AppVersion{}
	}

	out.AppVersionTypes = "helm"
	out.Isv = in.GetWorkspace()

	out.ClusterTotal = &rlsCount

	//out.URL = in.Spec.Url
	return out
}

func filterAppVersionByState(versions []*v1beta1.FederatedHelmApplicationVersion, states []string) []*v1beta1.FederatedHelmApplicationVersion {
	if len(states) == 0 {
		return versions
	}

	var j = 0
	for i := 0; i < len(versions); i++ {
		state := versions[i].Spec.Template.AuditSpec.State
		//default value is draft
		if state == "" {
			state = constants.StateDraft
		}
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

func convertAppVersion(in *v1beta1.FederatedHelmApplicationVersion) *AppVersion {
	if in == nil {
		return nil
	}
	out := AppVersion{}
	out.AppId = in.GetHelmApplicationId()
	out.Active = true
	t := in.CreationTimestamp.Time
	date := strfmt.DateTime(t)
	out.CreateTime = &date
	template := in.Spec.Template
	if len(template.AuditSpec.Audit) > 0 {
		t = template.AuditSpec.Audit[0].Time.Time
		updateDate := strfmt.DateTime(t)
		out.UpdateTime = &updateDate
	} else {
		out.UpdateTime = &date
	}
	if template.Spec.Metadata != nil {
		out.Description = template.Spec.Description
		out.Icon = template.Spec.Icon
	}

	out.Status = template.AuditSpec.State
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

	//TODO, add credential
	//if in.Credential != nil {
	//	out.Credential = in.Spec.Credential
	//}

	out.Status = "active"
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

type FederatedHelmApplicationList []*v1beta1.FederatedHelmApplication

func (l FederatedHelmApplicationList) Len() int      { return len(l) }
func (l FederatedHelmApplicationList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l FederatedHelmApplicationList) Less(i, j int) bool {
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

type AppVersions []*v1beta1.FederatedHelmApplicationVersion

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

//
//type sortable interface {
//	GetTrueName() string
//	metav1.ObjectMetaAccessor
//}
//
//type sortableList []sortable
//func (l sortableList) Len() int      { return len(l) }
//func (l sortableList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
//func (l sortableList) Less(i, j int) bool {
//	m1 := l[i].GetObjectMeta()
//	m2 := l[i].GetObjectMeta()
//	t1 := m1.GetCreationTimestamp().UnixNano()
//	t2 := m2.GetCreationTimestamp().UnixNano()
//	if t1 < t2 {
//		return true
//	} else if t1 > t2 {
//		return false
//	} else {
//		n1 := l[i].GetTrueName()
//		n2 := l[j].GetTrueName()
//		return n1 < n2
//	}
//}

//buildApplicationVersion  build an application version
//packageData base64 encoded package data
func buildApplicationVersion(app *v1beta1.FederatedHelmApplication, chrt helmrepoindex.VersionInterface, chartPackage *string, creator string) *v1beta1.FederatedHelmApplicationVersion {
	//create helm application version resource
	t := metav1.Now()
	ver := &v1alpha1.HelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				constants.CreatorAnnotationKey: creator,
			},
			Name: idutils.GetUuid36(constants.HelmApplicationVersionIdPrefix),
			Labels: map[string]string{
				constants.ChartApplicationIdLabelKey: app.GetHelmApplicationId(),
				constants.WorkspaceLabelKey:          app.GetWorkspace(),
				//constants.ChartVersionLabelKey:       chrt.GetVersionName(),
				//constants.ChartVersionLabelKey: chrt.GetVersion(),
			},
			//TODO, add OwnerReferences
			//OwnerReferences: []metav1.OwnerReference{
			//	{
			//		APIVersion: app.APIVersion,
			//		Kind:       app.Kind,
			//		Name:       app.GetName(),
			//		UID:        app.UID,
			//	},
			//},
		},
		Spec: v1alpha1.HelmApplicationVersionSpec{
			Metadata: &v1alpha1.Metadata{
				Version:     chrt.GetVersion(),
				AppVersion:  chrt.GetAppVersion(),
				Name:        chrt.GetName(),
				Icon:        chrt.GetIcon(),
				Home:        chrt.GetHome(),
				Description: strings2.ShortenString(chrt.GetDescription(), constants.MsgLen),
			},
			Created: &t,
			Data:    strfmt.Base64(*chartPackage),
		},
	}

	return toFederateHelmApplicationVersion(ver)
}

func buildApplicationVersionAudit(appVersion *v1beta1.FederatedHelmApplicationVersion) *v1alpha1.HelmAudit {
	t := metav1.Now()
	audit := &v1alpha1.HelmAudit{
		ObjectMeta: metav1.ObjectMeta{
			Name: appVersion.GetName(),
			Labels: map[string]string{
				constants.ChartApplicationIdLabelKey: appVersion.GetHelmApplicationId(),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1beta1.SchemeGroupVersion.String(),
					Kind:       v1beta1.FederatedHelmApplicationVersionKind,
					Name:       appVersion.GetName(),
					UID:        appVersion.UID,
				},
			},
		},
		Spec: v1alpha1.HelmAuditSpec{
			State: constants.StateDraft,
			Audit: []v1alpha1.Audit{
				{
					State:    constants.StateDraft,
					Time:     t,
					Operator: appVersion.GetCreator(),
				},
			},
		},
	}

	return audit
}

func convertAppVersionReview(appVersion *v1beta1.FederatedHelmApplicationVersion) *AppVersionReview {
	review := &AppVersionReview{}

	review.Reviewer = appVersion.Spec.Template.AuditSpec.Audit[0].Operator
	review.ReviewId = appVersion.Spec.Template.AuditSpec.Audit[0].Operator
	review.Status = appVersion.Spec.Template.AuditSpec.State
	review.AppId = appVersion.GetHelmApplicationId()
	review.VersionID = appVersion.GetHelmApplicationVersionId()
	review.Phase = AppVersionReviewPhaseOAIGen{}
	review.VersionName = appVersion.GetVersionName()

	review.StatusTime = strfmt.DateTime(appVersion.Spec.Template.AuditSpec.Audit[0].Time.Time)
	review.AppName = appVersion.Name
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
