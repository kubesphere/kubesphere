// /*
// Copyright 2020 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
//

package reposcache

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"os"
	"path"
	"strings"
	"sync"
)

var WorkDir string

func NewReposCache() ReposCache {
	return &cachedRepos{
		chartsInRepo:  map[workspace]map[string]int{},
		repos:         map[string]*v1alpha1.HelmRepo{},
		apps:          map[string]*v1alpha1.HelmApplication{},
		versions:      map[string]*v1alpha1.HelmApplicationVersion{},
		repoCtgCounts: map[string]map[string]int{},
	}
}

type ReposCache interface {
	AddRepo(repo *v1alpha1.HelmRepo) error
	DeleteRepo(repo *v1alpha1.HelmRepo) error

	GetApplication(string) (*v1alpha1.HelmApplication, bool)
	GetAppVersion(string) (*v1alpha1.HelmApplicationVersion, bool, error)
	GetAppVersionWithData(string) (*v1alpha1.HelmApplicationVersion, bool, error)

	ListAppVersionsByAppId(appId string) (ret []*v1alpha1.HelmApplicationVersion, exists bool)
	ListApplicationsByRepoId(repoId string) (ret []*v1alpha1.HelmApplication, exists bool)
}

type workspace string
type cachedRepos struct {
	sync.RWMutex

	chartsInRepo  map[workspace]map[string]int
	repoCtgCounts map[string]map[string]int

	repos    map[string]*v1alpha1.HelmRepo
	apps     map[string]*v1alpha1.HelmApplication
	versions map[string]*v1alpha1.HelmApplicationVersion
}

func (c *cachedRepos) deleteRepo(repo *v1alpha1.HelmRepo) {
	if len(repo.Status.Data) == 0 {
		return
	}
	index, err := helmrepoindex.ByteArrayToSavedIndex([]byte(repo.Status.Data))
	if err != nil {
		klog.Errorf("json unmarshal repo %s failed, error: %s", repo.Name, err)
		return
	}

	klog.V(4).Infof("delete repo %s from cache", repo.Name)
	c.Lock()
	defer c.Unlock()
	repoId := repo.GetHelmRepoId()
	ws := workspace(repo.GetWorkspace())
	if _, exists := c.chartsInRepo[ws]; exists {
		delete(c.chartsInRepo[ws], repoId)
	}

	delete(c.repoCtgCounts, repoId)
	delete(c.repos, repoId)

	for _, app := range index.Applications {
		delete(c.apps, app.ApplicationId)
		for _, ver := range app.Charts {
			delete(c.versions, ver.ApplicationVersionId)
		}
	}
}

func loadBuiltinChartData(name, version string) ([]byte, error) {
	fName := path.Join(WorkDir, "chart", fmt.Sprintf("%s-%s.tgz", name, version))
	f, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		klog.Errorf("read index failed, error: %s", err)
		return nil, err
	}
	return data, nil
}

func (c *cachedRepos) DeleteRepo(repo *v1alpha1.HelmRepo) error {
	c.deleteRepo(repo)
	return nil
}

func (c *cachedRepos) GetApplication(appId string) (app *v1alpha1.HelmApplication, exists bool) {
	c.RLock()
	defer c.RUnlock()
	if app, exists := c.apps[appId]; exists {
		return app, true
	}
	return
}

func (c *cachedRepos) AddRepo(repo *v1alpha1.HelmRepo) error {
	return c.addRepo(repo, false)
}

//Add new Repo to cachedRepos
func (c *cachedRepos) addRepo(repo *v1alpha1.HelmRepo, builtin bool) error {
	if len(repo.Status.Data) == 0 {
		return nil
	}
	index, err := helmrepoindex.ByteArrayToSavedIndex([]byte(repo.Status.Data))
	if err != nil {
		klog.Errorf("json unmarshal repo %s failed, error: %s", repo.Name, err)
		return err
	}

	klog.V(4).Infof("add repo %s to cache", repo.Name)

	c.Lock()
	defer c.Unlock()

	ws := workspace(repo.GetWorkspace())
	if _, exists := c.chartsInRepo[ws]; !exists {
		c.chartsInRepo[ws] = make(map[string]int)
	}

	repoId := repo.GetHelmRepoId()
	c.repos[repoId] = repo
	//c.repoCtgCounts[repo.GetHelmRepoId()] = make(map[string]int)
	if _, exists := c.repoCtgCounts[repoId]; !exists {
		c.repoCtgCounts[repoId] = map[string]int{}
	}
	var appName string

	chartsCount := 0
	for key, app := range index.Applications {
		if builtin {
			appName = v1alpha1.HelmApplicationIdPrefix + app.Name
		} else {
			appName = app.ApplicationId
		}

		HelmApp := v1alpha1.HelmApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name: appName,
				Annotations: map[string]string{
					constants.CreatorAnnotationKey: repo.GetCreator(),
				},
				Labels: map[string]string{
					constants.ChartRepoIdLabelKey: repo.GetHelmRepoId(),
				},
			},
			Spec: v1alpha1.HelmApplicationSpec{
				Name:        key,
				Description: app.Description,
				Icon:        app.Icon,
			},
			Status: v1alpha1.HelmApplicationStatus{
				State: v1alpha1.StateActive,
			},
		}
		c.apps[app.ApplicationId] = &HelmApp

		var ctg, appVerName string
		var chartData []byte
		for _, ver := range app.Charts {
			chartsCount += 1
			if ver.Annotations != nil && ver.Annotations["category"] != "" {
				ctg = ver.Annotations["category"]
			}
			if builtin {
				appVerName = base64.StdEncoding.EncodeToString([]byte(ver.Name + ver.Version))
				chartData, err = loadBuiltinChartData(ver.Name, ver.Version)
				if err != nil {
					return err
				}
			} else {
				appVerName = ver.ApplicationVersionId
			}

			version := &v1alpha1.HelmApplicationVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appVerName,
					Annotations: map[string]string{constants.CreatorAnnotationKey: repo.GetCreator()},
					Labels: map[string]string{
						constants.ChartApplicationIdLabelKey: appName,
						constants.ChartRepoIdLabelKey:        repo.GetHelmRepoId(),
					},
					CreationTimestamp: metav1.Time{Time: ver.Created},
				},
				Spec: v1alpha1.HelmApplicationVersionSpec{
					Metadata: &v1alpha1.Metadata{
						Name:       ver.Name,
						AppVersion: ver.AppVersion,
						Version:    ver.Version,
					},
					URLs:   ver.URLs,
					Digest: ver.Digest,
					Data:   chartData,
				},
				Status: v1alpha1.HelmApplicationVersionStatus{
					State: v1alpha1.StateActive,
				},
			}
			c.versions[ver.ApplicationVersionId] = version
		}

		//modify application category
		ctgId := ""
		if ctg != "" {
			if c.apps[app.ApplicationId].Annotations == nil {
				c.apps[app.ApplicationId].Annotations = map[string]string{constants.CategoryIdLabelKey: ctg}
			} else {
				c.apps[app.ApplicationId].Annotations[constants.CategoryIdLabelKey] = ctg
			}
			ctgId = ctg
		} else {
			ctgId = v1alpha1.UncategorizedId
		}

		if _, exists := c.repoCtgCounts[repoId][ctgId]; !exists {
			c.repoCtgCounts[repoId][ctgId] = 1
		} else {
			c.repoCtgCounts[repoId][ctgId] += 1
		}
	}

	c.chartsInRepo[ws][repo.GetHelmRepoId()] = chartsCount

	return nil
}

func (c *cachedRepos) ListApplicationsByRepoId(repoId string) (ret []*v1alpha1.HelmApplication, exists bool) {
	c.RLock()
	defer c.RUnlock()

	if repo, exists := c.repos[repoId]; !exists {
		return nil, false
	} else {
		ret = make([]*v1alpha1.HelmApplication, 0, 10)
		for _, app := range c.apps {
			if app.GetHelmRepoId() == repo.Name {
				ret = append(ret, app)
			}
		}
	}
	return ret, true
}

func (c *cachedRepos) ListAppVersionsByAppId(appId string) (ret []*v1alpha1.HelmApplicationVersion, exists bool) {
	c.RLock()
	defer c.RUnlock()

	if _, exists := c.apps[appId]; !exists {
		return nil, false
	}

	ret = make([]*v1alpha1.HelmApplicationVersion, 0, 10)
	for _, ver := range c.versions {
		if ver.GetHelmApplicationId() == appId {
			ret = append(ret, ver)
		}
	}
	return ret, true
}

func (c *cachedRepos) getAppVersion(versionId string, withData bool) (ret *v1alpha1.HelmApplicationVersion, exists bool, err error) {
	c.RLock()
	if version, exists := c.versions[versionId]; exists {
		//builtin chart data
		if withData {
			if len(version.Spec.Data) != 0 {
				c.RUnlock()
				return version, true, nil
			}

			if len(version.Spec.URLs) == 0 {
				c.RUnlock()
				return nil, true, errors.New("invalid chart spec")
			}
			var repo *v1alpha1.HelmRepo
			var exists bool
			if repo, exists = c.repos[version.GetHelmRepoId()]; !exists {
				c.RUnlock()
				klog.Errorf("load repo for app version: %s/%s failed",
					version.GetWorkspace(), version.GetTrueName())
				return nil, true, err
			}

			c.RUnlock()
			url := version.Spec.URLs[0]
			if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "s3://")) {
				url = repo.Spec.Url + "/" + url
			}

			buf, err := helmrepoindex.LoadChart(context.TODO(), url, &repo.Spec.Credential)
			if err != nil {
				klog.Errorf("load chart data for app version: %s/%s failed, error : %s", version.GetTrueName(),
					version.GetTrueName(), err)
				return nil, true, err
			}
			version.Spec.Data = buf.Bytes()
			return version, true, nil
		} else {
			c.RUnlock()
			return version, true, nil
		}
	} else {
		c.RUnlock()
		//version does not exists
		return nil, false, nil
	}
}
func (c *cachedRepos) GetAppVersion(versionId string) (ret *v1alpha1.HelmApplicationVersion, exists bool, err error) {
	return c.getAppVersion(versionId, false)
}

func (c *cachedRepos) GetAppVersionWithData(versionId string) (ret *v1alpha1.HelmApplicationVersion, exists bool, err error) {
	return c.getAppVersion(versionId, true)
}
