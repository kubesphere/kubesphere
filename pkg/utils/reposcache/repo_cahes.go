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
	"errors"
	"strings"
	"sync"

	"k8s.io/client-go/tools/cache"

	"github.com/Masterminds/semver/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"

	"kubesphere.io/api/application/v1alpha1"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
)

const (
	CategoryIndexer       = "category_indexer"
	CategoryAnnotationKey = "app.kubesphere.io/category"
)

var WorkDir string

func NewReposCache() ReposCache {
	return &cachedRepos{
		chartsInRepo:          map[workspace]map[string]int{},
		repos:                 map[string]*v1alpha1.HelmRepo{},
		apps:                  map[string]*v1alpha1.HelmApplication{},
		versions:              map[string]*v1alpha1.HelmApplicationVersion{},
		builtinCategoryCounts: map[string]int{},
	}
}

type ReposCache interface {
	AddRepo(repo *v1alpha1.HelmRepo) error
	DeleteRepo(repo *v1alpha1.HelmRepo) error
	UpdateRepo(old, new *v1alpha1.HelmRepo) error

	GetApplication(string) (*v1alpha1.HelmApplication, bool)
	GetAppVersion(string) (*v1alpha1.HelmApplicationVersion, bool, error)
	GetAppVersionWithData(string) (*v1alpha1.HelmApplicationVersion, bool, error)

	ListAppVersionsByAppId(appId string) (ret []*v1alpha1.HelmApplicationVersion, exists bool)
	ListApplicationsInRepo(repoId string) (ret []*v1alpha1.HelmApplication, exists bool)
	ListApplicationsInBuiltinRepo(selector labels.Selector) (ret []*v1alpha1.HelmApplication, exists bool)

	SetCategoryIndexer(indexer cache.Indexer)
	CopyCategoryCount() map[string]int
}

type workspace string
type cachedRepos struct {
	sync.RWMutex

	chartsInRepo map[workspace]map[string]int

	// builtinCategoryCounts saves the count of every category in the built-in repo.
	builtinCategoryCounts map[string]int

	repos    map[string]*v1alpha1.HelmRepo
	apps     map[string]*v1alpha1.HelmApplication
	versions map[string]*v1alpha1.HelmApplicationVersion

	// indexerOfHelmCtg is the indexer of HelmCategory, used to query the category id from category name.
	indexerOfHelmCtg cache.Indexer
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

	klog.V(2).Infof("delete repo %s from cache", repo.Name)

	repoId := repo.GetHelmRepoId()
	ws := workspace(repo.GetWorkspace())
	if _, exists := c.chartsInRepo[ws]; exists {
		delete(c.chartsInRepo[ws], repoId)
	}

	delete(c.repos, repoId)

	for _, app := range index.Applications {
		if _, exists := c.apps[app.ApplicationId]; !exists {
			continue
		}
		if helmrepoindex.IsBuiltInRepo(repo.Name) {
			ctgId := c.apps[app.ApplicationId].Labels[constants.CategoryIdLabelKey]
			if ctgId != "" {
				c.builtinCategoryCounts[ctgId] -= 1
			}
		}

		delete(c.apps, app.ApplicationId)
		for _, ver := range app.Charts {
			delete(c.versions, ver.ApplicationVersionId)
		}
	}
}

func (c *cachedRepos) DeleteRepo(repo *v1alpha1.HelmRepo) error {
	c.Lock()
	defer c.Unlock()

	c.deleteRepo(repo)
	return nil
}

// CopyCategoryCount copies the internal map to avoid `concurrent map iteration and map write`.
func (c *cachedRepos) CopyCategoryCount() map[string]int {
	c.RLock()
	defer c.RUnlock()

	ret := make(map[string]int, len(c.builtinCategoryCounts))
	for k, v := range c.builtinCategoryCounts {
		ret[k] = v
	}

	return ret
}

func (c *cachedRepos) SetCategoryIndexer(indexer cache.Indexer) {
	c.Lock()
	c.indexerOfHelmCtg = indexer
	c.Unlock()
}

// translateCategoryNameToId translate a category-name to a category-id.
// The caller should hold the lock
func (c *cachedRepos) translateCategoryNameToId(ctgName string) string {
	if c.indexerOfHelmCtg == nil || ctgName == "" {
		return v1alpha1.UncategorizedId
	}

	if items, err := c.indexerOfHelmCtg.ByIndex(CategoryIndexer, ctgName); len(items) == 0 || err != nil {
		return v1alpha1.UncategorizedId
	} else {
		obj, _ := items[0].(*v1alpha1.HelmCategory)
		return obj.Name
	}
}

func (c *cachedRepos) GetApplication(appId string) (app *v1alpha1.HelmApplication, exists bool) {
	c.RLock()
	defer c.RUnlock()
	if app, exists := c.apps[appId]; exists {
		return app, true
	}
	return
}

func (c *cachedRepos) UpdateRepo(old, new *v1alpha1.HelmRepo) error {
	if old.Status.Data == new.Status.Data {
		return nil
	}
	c.Lock()
	defer c.Unlock()

	c.deleteRepo(old)
	return c.addRepo(new, false)
}

func (c *cachedRepos) AddRepo(repo *v1alpha1.HelmRepo) error {
	c.Lock()
	defer c.Unlock()
	return c.addRepo(repo, false)
}

// Add a new Repo to cachedRepos
func (c *cachedRepos) addRepo(repo *v1alpha1.HelmRepo, builtin bool) error {
	if len(repo.Status.Data) == 0 {
		return nil
	}
	index, err := helmrepoindex.ByteArrayToSavedIndex([]byte(repo.Status.Data))
	if err != nil {
		klog.Errorf("json unmarshal repo %s failed, error: %s", repo.Name, err)
		return err
	}

	klog.V(2).Infof("add repo %s to cache", repo.Name)

	ws := workspace(repo.GetWorkspace())
	if _, exists := c.chartsInRepo[ws]; !exists {
		c.chartsInRepo[ws] = make(map[string]int)
	}

	repoId := repo.GetHelmRepoId()
	c.repos[repoId] = repo
	var appName string

	chartsCount := 0
	for key, app := range index.Applications {
		appName = app.ApplicationId

		appLabels := make(map[string]string)
		if helmrepoindex.IsBuiltInRepo(repo.Name) {
			appLabels[constants.WorkspaceLabelKey] = "system-workspace"
		}

		appLabels[constants.ChartRepoIdLabelKey] = repoId

		helmApp := v1alpha1.HelmApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name: appName,
				Annotations: map[string]string{
					constants.CreatorAnnotationKey: repo.GetCreator(),
				},
				Labels:            appLabels,
				CreationTimestamp: metav1.Time{Time: app.Created},
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
		c.apps[app.ApplicationId] = &helmApp

		var ctg, appVerName string
		var chartData []byte
		var latestVersionName string
		var latestSemver *semver.Version

		// build all the versions of this app
		for _, chartVersion := range app.Charts {
			chartsCount += 1
			hvw := helmrepoindex.HelmVersionWrapper{ChartVersion: &chartVersion.ChartVersion}
			appVerName = chartVersion.ApplicationVersionId
			version := &v1alpha1.HelmApplicationVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:        appVerName,
					Annotations: map[string]string{constants.CreatorAnnotationKey: repo.GetCreator()},
					Labels: map[string]string{
						constants.ChartApplicationIdLabelKey: appName,
						constants.ChartRepoIdLabelKey:        repo.GetHelmRepoId(),
					},
					CreationTimestamp: metav1.Time{Time: chartVersion.Created},
				},
				Spec: v1alpha1.HelmApplicationVersionSpec{
					Metadata: &v1alpha1.Metadata{
						Name:       hvw.GetName(),
						AppVersion: hvw.GetAppVersion(),
						Version:    hvw.GetVersion(),
					},
					URLs:   chartVersion.URLs,
					Digest: chartVersion.Digest,
					Data:   chartData,
				},
				Status: v1alpha1.HelmApplicationVersionStatus{
					State: v1alpha1.StateActive,
				},
			}

			// It is not necessary to store these pieces of information when this is not a built-in repo.
			if helmrepoindex.IsBuiltInRepo(repo.Name) {
				version.Spec.Sources = hvw.GetRawSources()
				version.Spec.Maintainers = hvw.GetRawMaintainers()
				version.Spec.Home = hvw.GetHome()
			}
			c.versions[chartVersion.ApplicationVersionId] = version

			// Find the latest version.
			currSemver, err := semver.NewVersion(version.GetSemver())
			if err == nil {
				if latestSemver == nil {
					// the first valid semver
					latestSemver = currSemver
					latestVersionName = version.GetVersionName()

					// Use the category of the latest version as the category of the app.
					ctg = chartVersion.Annotations[CategoryAnnotationKey]
				} else if latestSemver.LessThan(currSemver) {
					// find a newer valid semver
					latestSemver = currSemver
					latestVersionName = version.GetVersionName()
					ctg = chartVersion.Annotations[CategoryAnnotationKey]
				}
			} else {
				// If the semver is invalid, just ignore it.
				klog.V(2).Infof("parse version failed, id: %s, err: %s", version.Name, err)
			}
		}

		helmApp.Status.LatestVersion = latestVersionName

		if helmrepoindex.IsBuiltInRepo(repo.Name) {
			// Add category id to the apps in the built-in repo
			ctgId := c.translateCategoryNameToId(ctg)
			if helmApp.Labels == nil {
				helmApp.Labels = map[string]string{}
			}
			helmApp.Labels[constants.CategoryIdLabelKey] = ctgId

			c.builtinCategoryCounts[ctgId] += 1
		}
	}

	c.chartsInRepo[ws][repo.GetHelmRepoId()] = chartsCount

	return nil
}

func (c *cachedRepos) ListApplicationsInRepo(repoId string) (ret []*v1alpha1.HelmApplication, exists bool) {
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

func (c *cachedRepos) ListApplicationsInBuiltinRepo(selector labels.Selector) (ret []*v1alpha1.HelmApplication, exists bool) {
	c.RLock()
	defer c.RUnlock()

	ret = make([]*v1alpha1.HelmApplication, 0, 20)
	for _, app := range c.apps {
		if strings.HasPrefix(app.GetHelmRepoId(), v1alpha1.BuiltinRepoPrefix) {
			if selector != nil && !selector.Empty() &&
				(app.Labels == nil || !selector.Matches(labels.Set(app.Labels))) { // If the selector is not empty, we must check whether the labels of the app match the selector.
				continue
			}
			ret = append(ret, app)
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
