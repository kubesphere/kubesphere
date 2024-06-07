/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"fmt"
	"net/url"

	"kubesphere.io/kubesphere/pkg/simple/client/application"

	"kubesphere.io/kubesphere/pkg/api"

	"github.com/emicklei/go-restful/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

func (h *appHandler) CreateOrUpdateRepo(req *restful.Request, resp *restful.Response) {

	repoRequest := &appv2.Repo{}
	err := req.ReadEntity(repoRequest)
	if requestDone(err, resp) {
		return
	}
	if repoRequest.Name == appv2.UploadRepoKey {
		api.HandleBadRequest(resp, req, fmt.Errorf("repo name %s is not allowed", appv2.UploadRepoKey))
		return
	}
	repoId := req.PathParameter("repo")
	if repoId == "" {
		repoId = repoRequest.Name
	}

	parsedUrl, err := url.Parse(repoRequest.Spec.Url)
	if requestDone(err, resp) {
		return
	}

	if parsedUrl.User != nil {
		repoRequest.Spec.Credential.Username = parsedUrl.User.Username()
		repoRequest.Spec.Credential.Password, _ = parsedUrl.User.Password()
	}

	_, err = application.LoadRepoIndex(repoRequest.Spec.Url, repoRequest.Spec.Credential)
	if requestDone(err, resp) {
		return
	}

	if req.QueryParameter("validate") != "" {
		data := map[string]any{"ok": true}
		resp.WriteAsJson(data)
		return
	}

	repo := &appv2.Repo{}
	repo.Name = repoId
	if h.conflictedDone(req, resp, "repo", repo) {
		return
	}

	mutateFn := func() error {
		repo.Spec = appv2.RepoSpec{
			Url:         parsedUrl.String(),
			SyncPeriod:  repoRequest.Spec.SyncPeriod,
			Description: stringutils.ShortenString(repoRequest.Spec.Description, 512),
		}
		if parsedUrl.User != nil {
			repo.Spec.Credential.Username = parsedUrl.User.Username()
			repo.Spec.Credential.Password, _ = parsedUrl.User.Password()
		}
		if repo.GetLabels() == nil {
			repo.SetLabels(map[string]string{})
		}
		repo.Labels[constants.WorkspaceLabelKey] = repoRequest.Labels[constants.WorkspaceLabelKey]

		if repo.GetAnnotations() == nil {
			repo.SetAnnotations(map[string]string{})
		}
		ant := repoRequest.GetAnnotations()
		repo.Annotations[constants.DisplayNameAnnotationKey] = ant[constants.DisplayNameAnnotationKey]

		return nil
	}

	_, err = controllerutil.CreateOrUpdate(req.Request.Context(), h.client, repo, mutateFn)
	if requestDone(err, resp) {
		return
	}
	data := map[string]interface{}{"repo_id": repoId}

	resp.WriteAsJson(data)
}

func (h *appHandler) DeleteRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	err := h.client.Delete(req.Request.Context(), &appv2.Repo{ObjectMeta: metav1.ObjectMeta{Name: repoId}})
	if requestDone(err, resp) {
		return
	}
	klog.V(4).Info("delete repo: ", repoId)

	resp.WriteEntity(errors.None)
}

func (h *appHandler) DescribeRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	key := runtimeclient.ObjectKey{Name: repoId}
	repo := &appv2.Repo{}
	err := h.client.Get(req.Request.Context(), key, repo)
	if requestDone(err, resp) {
		return
	}
	repo.SetManagedFields(nil)

	resp.WriteEntity(repo)
}

func (h *appHandler) ListRepos(req *restful.Request, resp *restful.Response) {

	helmRepoList := &appv2.RepoList{}
	err := h.client.List(req.Request.Context(), helmRepoList)
	if requestDone(err, resp) {
		return
	}
	workspace := req.PathParameter("workspace")
	filteredList := &appv2.RepoList{}
	for _, repo := range helmRepoList.Items {
		allowList := []string{appv2.SystemWorkspace, workspace}
		if !stringutils.StringIn(repo.Labels[constants.WorkspaceLabelKey], allowList) {
			continue
		}
		filteredList.Items = append(filteredList.Items, repo)
	}

	resp.WriteEntity(convertToListResult(filteredList, req))
}

func (h *appHandler) ListRepoEvents(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	list := v1.EventList{}
	selector := fields.SelectorFromSet(fields.Set{
		"involvedObject.name": repoId,
	})

	opt := &runtimeclient.ListOptions{FieldSelector: selector}
	err := h.client.List(req.Request.Context(), &list, opt)
	if requestDone(err, resp) {
		return
	}

	resp.WriteEntity(convertToListResult(&list, req))
}
