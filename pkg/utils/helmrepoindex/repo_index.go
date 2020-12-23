package helmrepoindex

import (
	"context"
	"helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"net/url"
	"path"
	"sigs.k8s.io/yaml"
	"time"
)

//copy from helm
func LoadRepoIndex(ctx context.Context, u string, cred *v1alpha1.HelmRepoCredential) (*helmrepo.IndexFile, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	parsedURL.RawPath = path.Join(parsedURL.RawPath, "index.yaml")
	parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")
	//
	//skipTLS := true
	//if cred.InsecureSkipTLSVerify != nil && !*cred.InsecureSkipTLSVerify {
	//	skipTLS = false
	//}

	indexURL := parsedURL.String()
	// TODO add user-agent
	g, _ := getter.NewHTTPGetter()
	resp, err := g.Get(indexURL,
		//getter.WithTimeout(5*time.Minute),
		getter.WithURL(u),
		//getter.WithInsecureSkipVerifyTLS(skipTLS),
		getter.WithTLSClientConfig(cred.CertFile, cred.KeyFile, cred.CAFile),
		getter.WithBasicAuth(cred.Username, cred.Password),
	)
	index, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	indexFile, err := loadIndex(index)
	if err != nil {
		return nil, err
	}

	return indexFile, nil
}

// loadIndex loads an index file and does minimal validity checking.
//
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (*helmrepo.IndexFile, error) {
	i := &helmrepo.IndexFile{}
	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return i, err
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return i, helmrepo.ErrNoAPIVersion
	}
	return i, nil
}

//merge new index with index from crd
func MergeRepoIndex(index *helmrepo.IndexFile, existsSavedIndex *SavedIndex) *SavedIndex {
	saved := &SavedIndex{}
	if index == nil {
		return existsSavedIndex
	}

	saved.Applications = make(map[string]*Application)
	for name := range existsSavedIndex.Applications {
		saved.Applications[name] = existsSavedIndex.Applications[name]
	}

	//just copy fields from index
	//saved.ServerInfo = index.ServerInfo
	saved.APIVersion = index.APIVersion
	saved.Generated = index.Generated
	saved.PublicKeys = index.PublicKeys
	//saved.Annotations = index.Annotations

	allNames := make(map[string]bool, len(index.Entries))
	for name, versions := range index.Entries {
		//add new applications
		if _, exists := saved.Applications[name]; !exists {
			application := Application{
				Name:          name,
				ApplicationId: idutils.GetUuid36(constants.HelmApplicationIdPrefix),
				Description:   versions[0].Description,
				Icon:          versions[0].Icon,
				Status:        "active",
			}

			charts := make([]*ChartVersion, 0, len(versions))
			for ind := range versions {
				chart := &ChartVersion{
					ApplicationId:        application.ApplicationId,
					ApplicationVersionId: idutils.GetUuid36(constants.HelmApplicationVersionIdPrefix),
					ChartVersion:         *versions[ind],
				}
				charts = append(charts, chart)
			}
			application.Charts = charts
			saved.Applications[name] = &application
		} else {
			//todo update exists applications
		}
		allNames[name] = true
	}

	for name := range saved.Applications {
		if _, exists := allNames[name]; !exists {
			//delete
			delete(saved.Applications, name)
		}
	}

	return saved
}

func (i *SavedIndex) GetApplicationVersion(appId, versionId string) *v1alpha1.HelmApplicationVersion {
	for _, app := range i.Applications {
		if app.ApplicationId == appId {
			for _, ver := range app.Charts {
				if ver.ApplicationVersionId == versionId {
					version := &v1alpha1.HelmApplicationVersion{
						ObjectMeta: metav1.ObjectMeta{
							Name: versionId,
							Labels: map[string]string{
								constants.ChartApplicationIdLabelKey: appId,
								//constants.NameLabelKey:               ver.Name,
								//constants.ChartVersionLabelKey:       ver.Version,
							},
						},
						Spec: v1alpha1.HelmApplicationVersionSpec{
							URLs:   ver.URLs,
							Digest: ver.Digest,
							Metadata: &v1alpha1.Metadata{
								Name:       ver.Name,
								AppVersion: ver.AppVersion,
								Version:    ver.Version,
							},
						},
					}
					return version
				}
			}
		}
	}
	return nil
}

type SavedIndex struct {
	// This is used ONLY for validation against chartmuseum's index files and is discarded after validation.
	ServerInfo   map[string]interface{}  `json:"serverInfo,omitempty"`
	APIVersion   string                  `json:"apiVersion"`
	Generated    time.Time               `json:"generated"`
	Applications map[string]*Application `json:"apps"`
	PublicKeys   []string                `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// chart version with app id and app version id
type ChartVersion struct {
	//Do not save ApplicationId into crd
	ApplicationId         string `json:"-"`
	ApplicationVersionId  string `json:"verId"`
	helmrepo.ChartVersion `json:",inline"`
}

type Application struct {
	//application name
	Name          string `json:"name"`
	ApplicationId string `json:"appId"`
	//chart description
	Description string `json:"desc"`
	//application status
	Status string `json:"status"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`

	Charts []*ChartVersion `json:"charts"`
}
