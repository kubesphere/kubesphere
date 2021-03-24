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
	"context"
	"encoding/json"
	"github.com/go-openapi/strfmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	typed_v1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/utils/reposcache"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
)

const DescriptionLen = 512

type RepoInterface interface {
	CreateRepo(repo *v1alpha1.HelmRepo) (*CreateRepoResponse, error)
	DeleteRepo(id string) error
	ValidateRepo(u string, request *v1alpha1.HelmRepoCredential) (*ValidateRepoResponse, error)
	ModifyRepo(id string, request *ModifyRepoRequest) error
	DescribeRepo(id string) (*Repo, error)
	ListRepos(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	DoRepoAction(repoId string, request *RepoActionRequest) error
	ListRepoEvents(repoId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error)
}

type repoOperator struct {
	cachedRepos reposcache.ReposCache
	informers   externalversions.SharedInformerFactory
	repoClient  typed_v1alpha1.ApplicationV1alpha1Interface
	repoLister  listers_v1alpha1.HelmRepoLister
	rlsLister   listers_v1alpha1.HelmReleaseLister
}

func newRepoOperator(cachedRepos reposcache.ReposCache, informers externalversions.SharedInformerFactory, ksClient versioned.Interface) RepoInterface {
	return &repoOperator{
		cachedRepos: cachedRepos,
		informers:   informers,
		repoClient:  ksClient.ApplicationV1alpha1(),
		repoLister:  informers.Application().V1alpha1().HelmRepos().Lister(),
		rlsLister:   informers.Application().V1alpha1().HelmReleases().Lister(),
	}
}

// TODO implement DoRepoAction
func (c *repoOperator) DoRepoAction(repoId string, request *RepoActionRequest) error {
	repo, err := c.repoLister.Get(repoId)
	if err != nil {
		return err
	}
	if request.Workspace != repo.GetWorkspace() {
		return nil
	}

	patch := client.MergeFrom(repo)
	copyRepo := repo.DeepCopy()
	copyRepo.Spec.Version += 1
	data, err := patch.Data(copyRepo)
	if err != nil {
		klog.Errorf("create patch [%s] failed, error: %s", repoId, err)
		return err
	}
	repo, err = c.repoClient.HelmRepos().Patch(context.TODO(), repoId, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("patch repo [%s] failed, error: %s", repoId, err)
		return err
	}
	return nil
}

func (c *repoOperator) ValidateRepo(u string, cred *v1alpha1.HelmRepoCredential) (*ValidateRepoResponse, error) {
	_, err := helmrepoindex.LoadRepoIndex(context.TODO(), u, cred)

	if err != nil {
		return nil, err
	}
	return &ValidateRepoResponse{Ok: true}, nil
}

func (c *repoOperator) CreateRepo(repo *v1alpha1.HelmRepo) (*CreateRepoResponse, error) {
	name := repo.GetTrueName()

	items, err := c.repoLister.List(labels.SelectorFromSet(map[string]string{constants.WorkspaceLabelKey: repo.GetWorkspace()}))
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("list helm repo failed: %s", err)
		return nil, err
	}

	for _, exists := range items {
		if exists.GetTrueName() == name {
			klog.Error(repoItemExists, "name: ", name)
			return nil, repoItemExists
		}
	}

	repo.Spec.Description = stringutils.ShortenString(repo.Spec.Description, DescriptionLen)
	_, err = c.repoClient.HelmRepos().Create(context.TODO(), repo, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create helm repo failed, repo_id: %s, error: %s", repo.GetHelmRepoId(), err)
		return nil, err
	} else {
		klog.V(4).Infof("create helm repo success, repo_id: %s", repo.GetHelmRepoId())
	}

	return &CreateRepoResponse{repo.GetHelmRepoId()}, nil
}

func (c *repoOperator) DeleteRepo(id string) error {
	var err error
	err = c.repoClient.HelmRepos().Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("delete repo %s failed, error: %s", id, err)
		return err
	}
	klog.V(4).Infof("repo %s deleted", id)
	return nil
}

func (c *repoOperator) ModifyRepo(id string, request *ModifyRepoRequest) error {
	repo, err := c.repoClient.HelmRepos().Get(context.TODO(), id, metav1.GetOptions{})

	if err != nil {
		klog.Error("get repo failed", err)
		return err
	}

	repoCopy := repo.DeepCopy()
	if request.Description != nil {
		repoCopy.Spec.Description = stringutils.ShortenString(*request.Description, DescriptionLen)
	}

	if request.Name != nil && len(*request.Name) > 0 && *request.Name != repoCopy.Name {
		items, err := c.repoLister.List(labels.SelectorFromSet(map[string]string{constants.WorkspaceLabelKey: repo.GetWorkspace()}))
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("list helm repo failed: %s", err)
			return err
		}

		for _, exists := range items {
			if exists.GetTrueName() == *request.Name {
				klog.Error(repoItemExists, "name: ", *request.Name)
				return repoItemExists
			}
		}

		repoCopy.Spec.Name = *request.Name
	}

	// modify credential
	if request.URL != nil && len(*request.URL) > 0 {
		parsedUrl, err := url.Parse(*request.URL)
		if err != nil {
			return err
		}
		userInfo := parsedUrl.User
		// trim the credential from url
		parsedUrl.User = nil
		cred := &v1alpha1.HelmRepoCredential{}
		if strings.HasPrefix(*request.URL, "https://") || strings.HasPrefix(*request.URL, "http://") {
			if userInfo != nil {
				cred.Password, _ = userInfo.Password()
				cred.Username = userInfo.Username()
			} else {
				// trim the old credential
				cred.Password, _ = userInfo.Password()
				cred.Username = userInfo.Username()
			}
		} else if strings.HasPrefix(*request.URL, "s3://") {
			cfg := v1alpha1.S3Config{}
			err := json.Unmarshal([]byte(*request.Credential), &cfg)
			if err != nil {
				return err
			}
			cred.S3Config = cfg
		}

		repoCopy.Spec.Credential = *cred
		repoCopy.Spec.Url = parsedUrl.String()
	}

	patch := client.MergeFrom(repo)
	repoCopy.Spec.Version += 1
	data, err := patch.Data(repoCopy)
	if err != nil {
		klog.Error("create patch failed", err)
		return err
	}

	// data == "{}", need not to patch
	if len(data) == 2 {
		return nil
	}

	repo, err = c.repoClient.HelmRepos().Patch(context.TODO(), id, patch.Type(), data, metav1.PatchOptions{})

	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *repoOperator) DescribeRepo(id string) (*Repo, error) {
	repo, err := c.repoClient.HelmRepos().Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var desRepo Repo

	desRepo.URL = repo.Spec.Url
	desRepo.Description = repo.Spec.Description
	desRepo.Name = repo.GetTrueName()
	desRepo.RepoId = repo.Name
	dt, _ := strfmt.ParseDateTime(repo.CreationTimestamp.String())
	desRepo.CreateTime = &dt

	return &desRepo, nil
}

func (c *repoOperator) ListRepos(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	ls := labels.NewSelector()
	r, _ := labels.NewRequirement(constants.WorkspaceLabelKey, selection.Equals, []string{conditions.Match[WorkspaceLabel]})
	ls = ls.Add([]labels.Requirement{*r}...)

	repos, err := c.repoLister.List(ls)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if conditions.Match[Keyword] != "" {
		repos = helmRepoFilter(conditions.Match[Keyword], repos)
	}

	if reverse {
		sort.Sort(sort.Reverse(HelmRepoList(repos)))
	} else {
		sort.Sort(HelmRepoList(repos))
	}

	items := make([]interface{}, 0, limit)
	for i, j := offset, 0; i < len(repos) && j < limit; i, j = i+1, j+1 {
		items = append(items, convertRepo(repos[i]))
	}
	return &models.PageableResponse{Items: items, TotalCount: len(repos)}, nil
}

func helmRepoFilter(namePrefix string, list []*v1alpha1.HelmRepo) (res []*v1alpha1.HelmRepo) {
	for _, repo := range list {
		name := repo.GetTrueName()
		if strings.Contains(name, namePrefix) {
			res = append(res, repo)
		}
	}
	return
}

type HelmRepoList []*v1alpha1.HelmRepo

func (l HelmRepoList) Len() int      { return len(l) }
func (l HelmRepoList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l HelmRepoList) Less(i, j int) bool {
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

func (c *repoOperator) ListRepoEvents(repoId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error) {

	repo, err := c.repoClient.HelmRepos().Get(context.TODO(), repoId, metav1.GetOptions{})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0, limit)
	for i, j := offset, 0; i < len(repo.Status.SyncState) && j < limit; i, j = i+1, j+1 {
		items = append(items, convertRepoEvent(&repo.ObjectMeta, &repo.Status.SyncState[j]))
	}

	return &models.PageableResponse{Items: items, TotalCount: len(repo.Status.SyncState)}, nil
}
