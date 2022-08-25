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

package helmrepoindex

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	helmrepo "helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"kubesphere.io/api/application/v1alpha1"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
)

const IndexYaml = "index.yaml"

func LoadRepoIndex(ctx context.Context, u string, cred *v1alpha1.HelmRepoCredential) (*helmrepo.IndexFile, error) {

	if !strings.HasSuffix(u, "/") {
		u = fmt.Sprintf("%s/%s", u, IndexYaml)
	} else {
		u = fmt.Sprintf("%s%s", u, IndexYaml)
	}

	resp, err := loadData(ctx, u, cred)
	if err != nil {
		return nil, err
	}

	indexFile, err := loadIndex(resp.Bytes())
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
	if err := yaml.Unmarshal(data, i); err != nil {
		return i, err
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return i, helmrepo.ErrNoAPIVersion
	}
	return i, nil
}

var empty = struct{}{}

// MergeRepoIndex merge new index with index from crd
func MergeRepoIndex(repo *v1alpha1.HelmRepo, index *helmrepo.IndexFile, existsSavedIndex *SavedIndex) *SavedIndex {
	saved := &SavedIndex{}
	if index == nil {
		return existsSavedIndex
	}

	saved.Applications = make(map[string]*Application)
	if existsSavedIndex != nil {
		for name := range existsSavedIndex.Applications {
			saved.Applications[name] = existsSavedIndex.Applications[name]
		}
	}

	// just copy fields from index
	saved.APIVersion = index.APIVersion
	saved.Generated = index.Generated
	saved.PublicKeys = index.PublicKeys

	allAppNames := make(map[string]struct{}, len(index.Entries))
	for name, versions := range index.Entries {
		if len(versions) == 0 {
			continue
		}
		// add new applications
		if application, exists := saved.Applications[name]; !exists {
			application = &Application{
				Name:        name,
				Description: versions[0].Description,
				Icon:        versions[0].Icon,
				Created:     time.Now(),
			}

			// The app id will be added to the labels of the helm release.
			// But the apps in the repos which are created by the user may contain malformed text, so we generate a random name for them.
			// The apps in the system repo have been audited by the admin, so the name of the charts should not include malformed text.
			// Then we can add the name string to the labels of the k8s object.
			if IsBuiltInRepo(repo.Name) {
				application.ApplicationId = fmt.Sprintf("%s%s-%s", v1alpha1.HelmApplicationIdPrefix, repo.Name, name)
			} else {
				application.ApplicationId = idutils.GetUuid36(v1alpha1.HelmApplicationIdPrefix)
			}

			charts := make([]*ChartVersion, 0, len(versions))

			for ind := range versions {
				chart := &ChartVersion{
					ApplicationId: application.ApplicationId,
					ChartVersion:  *versions[ind],
				}

				chart.ApplicationVersionId = generateAppVersionId(repo, versions[ind].Name, versions[ind].Version)
				charts = append(charts, chart)

				// Use the creation time of the oldest chart as the creation time of the app.
				if versions[ind].Created.Before(application.Created) {
					application.Created = versions[ind].Created
				}
			}

			application.Charts = charts
			saved.Applications[name] = application
		} else {
			// update exists applications
			savedChartVersion := make(map[string]struct{})
			for _, ver := range application.Charts {
				savedChartVersion[ver.Version] = struct{}{}
			}
			charts := application.Charts
			var newVersion = make(map[string]struct{}, len(versions))
			for _, ver := range versions {
				// add new chart version
				if _, exists := savedChartVersion[ver.Version]; !exists {
					chart := &ChartVersion{
						ApplicationId: application.ApplicationId,
						ChartVersion:  *ver,
					}

					chart.ApplicationVersionId = generateAppVersionId(repo, ver.Name, ver.Version)
					charts = append(charts, chart)
				}
				newVersion[ver.Version] = empty
			}

			// delete not exists chart version
			for last, curr := 0, 0; curr < len(charts); {
				chart := charts[curr]
				version := chart.Version
				if _, exists := newVersion[version]; !exists {
					// version not exists, check next one
					curr++
				} else {
					// If last and curr point to the same place, there is nothing to do, just move to next.
					if last != curr {
						charts[last] = charts[curr]
					}
					last++
					curr++
				}
			}
			application.Charts = charts[:len(newVersion)]
			saved.Applications[name] = application
		}

		allAppNames[name] = empty
	}

	for name := range saved.Applications {
		if _, exists := allAppNames[name]; !exists {
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
							},
						},
						Spec: v1alpha1.HelmApplicationVersionSpec{
							URLs:   ver.URLs,
							Digest: ver.Digest,
							Metadata: &v1alpha1.Metadata{
								Name:        ver.Name,
								AppVersion:  ver.AppVersion,
								Version:     ver.Version,
								Annotations: ver.Annotations,
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

// The app version id will be added to the labels of the helm release.
// But the apps in the repos which are created by the user may contain malformed text, so we generate a random name for them.
// The apps in the system repo have been audited by the admin, so the name of the charts should not include malformed text.
// Then we can add the name string to the labels of the k8s object.
func generateAppVersionId(repo *v1alpha1.HelmRepo, chartName, version string) string {
	if IsBuiltInRepo(repo.Name) {
		return fmt.Sprintf("%s%s-%s-%s", v1alpha1.HelmApplicationIdPrefix, repo.Name, chartName, version)
	} else {
		return idutils.GetUuid36(v1alpha1.HelmApplicationVersionIdPrefix)
	}

}

// IsBuiltInRepo checks whether a repo is a built-in repo.
// All the built-in repos are located in the workspace system-workspace and the name starts with 'built-in'
// to differentiate from the repos created by the user.
func IsBuiltInRepo(repoName string) bool {
	return strings.HasPrefix(repoName, v1alpha1.BuiltinRepoPrefix)
}

type SavedIndex struct {
	APIVersion   string                  `json:"apiVersion"`
	Generated    time.Time               `json:"generated"`
	Applications map[string]*Application `json:"apps"`
	PublicKeys   []string                `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}

func ByteArrayToSavedIndex(data []byte) (*SavedIndex, error) {
	ret := &SavedIndex{}
	if len(data) == 0 {
		return ret, nil
	}

	enc := base64.URLEncoding
	buf := make([]byte, enc.DecodedLen(len(data)))
	n, err := enc.Decode(buf, data)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]

	r, err := zlib.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	r.Close()
	b, err := ioutil.ReadAll(r)

	if err != nil && err != io.EOF {
		return nil, err
	}

	err = json.Unmarshal(b, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *SavedIndex) Bytes() ([]byte, error) {

	d, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	w := zlib.NewWriter(buf)
	_, err = w.Write(d)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	encSrc := buf.Bytes()

	enc := base64.URLEncoding
	ret := make([]byte, enc.EncodedLen(len(encSrc)))

	enc.Encode(ret, encSrc)
	return ret, nil
}

// chart version with app id and app version id
type ChartVersion struct {
	// Do not save ApplicationId into crd
	ApplicationId         string `json:"-"`
	ApplicationVersionId  string `json:"verId"`
	helmrepo.ChartVersion `json:",inline"`
}

type Application struct {
	// application name
	Name          string `json:"name"`
	ApplicationId string `json:"appId"`
	// chart description
	Description string `json:"desc"`
	// application status
	Status string `json:"status"`
	// The URL to an icon file.
	Icon    string          `json:"icon,omitempty"`
	Created time.Time       `json:"created"`
	Charts  []*ChartVersion `json:"charts"`
}
