/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/git"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/quotas"
	"kubesphere.io/kubesphere/pkg/models/registries"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/models/revisions"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

type handler struct {
	componentsGetter    components.Getter
	resourceQuotaGetter quotas.ResourceQuotaGetter
	revisionGetter      revisions.RevisionGetter
	gitVerifier         git.GitVerifier
	registryGetter      registries.RegistryGetter
	kubeconfigOperator  kubeconfig.Interface
	resourceGetter      *resourcev1alpha3.Getter
}

func (h *handler) GetSystemHealthStatus(_ *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetComponents(_ *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetAllComponentsStatus()

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetClusterQuotas(_ *restful.Request, response *restful.Response) {
	result, err := h.resourceQuotaGetter.GetClusterQuota()
	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetNamespaceQuotas(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	quota, err := h.resourceQuotaGetter.GetNamespaceQuota(namespace)

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteAsJson(quota)
}

func (h *handler) GetDaemonSetRevision(request *restful.Request, response *restful.Response) {
	daemonset := request.PathParameter("daemonset")
	namespace := request.PathParameter("namespace")
	revision, err := strconv.Atoi(request.PathParameter("revision"))

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := h.revisionGetter.GetDaemonSetRevision(namespace, daemonset, revision)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetDeploymentRevision(request *restful.Request, response *restful.Response) {
	deploy := request.PathParameter("deployment")
	namespace := request.PathParameter("namespace")
	revision := request.PathParameter("revision")

	result, err := h.revisionGetter.GetDeploymentRevision(namespace, deploy, revision)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

func (h *handler) GetStatefulSetRevision(request *restful.Request, response *restful.Response) {
	statefulset := request.PathParameter("statefulset")
	namespace := request.PathParameter("namespace")
	revision, err := strconv.Atoi(request.PathParameter("revision"))
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := h.revisionGetter.GetStatefulSetRevision(namespace, statefulset, revision)
	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}
	response.WriteAsJson(result)
}

func (h *handler) VerifyGitCredential(request *restful.Request, response *restful.Response) {
	var credential api.GitCredential
	err := request.ReadEntity(&credential)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	var namespace, secretName string
	if credential.SecretRef != nil {
		namespace = credential.SecretRef.Namespace
		secretName = credential.SecretRef.Name
	}
	err = h.gitVerifier.VerifyGitCredential(credential.RemoteUrl, namespace, secretName)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	response.WriteAsJson(errors.None)
}

func (h *handler) VerifyRegistryCredential(request *restful.Request, response *restful.Response) {
	var credential api.RegistryCredential
	err := request.ReadEntity(&credential)
	if err != nil {
		api.HandleBadRequest(response, nil, err)
		return
	}

	err = h.registryGetter.VerifyRegistryCredential(credential)
	if err != nil {
		api.HandleBadRequest(response, nil, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h *handler) GetRegistryEntry(request *restful.Request, response *restful.Response) {
	imageName := request.QueryParameter("image")
	namespace := request.QueryParameter("namespace")
	secretName := request.QueryParameter("secret")
	insecure := request.QueryParameter("insecure") == "true"

	detail, err := h.registryGetter.GetEntry(namespace, secretName, imageName, insecure)
	if err != nil {
		api.HandleBadRequest(response, nil, err)
		return
	}

	response.WriteAsJson(detail)
}

func (h *handler) GetNamespacedAbnormalWorkloads(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	result := api.Workloads{
		Namespace: namespace,
		Count:     make(map[string]int),
	}

	for _, workloadType := range []string{api.ResourceKindDeployment, api.ResourceKindStatefulSet, api.ResourceKindDaemonSet, api.ResourceKindJob, api.ResourceKindPersistentVolumeClaim} {
		var notReadyStatus string

		switch workloadType {
		case api.ResourceKindPersistentVolumeClaim:
			notReadyStatus = strings.Join([]string{"pending", "lost"}, "|")
		case api.ResourceKindJob:
			notReadyStatus = "failed"
		default:
			notReadyStatus = "updating"
		}

		q := query.New()
		q.Filters[query.FieldStatus] = query.Value(notReadyStatus)

		res, err := h.resourceGetter.List(workloadType, namespace, q)
		if err != nil {
			api.HandleInternalError(response, nil, err)
		}

		result.Count[workloadType] = len(res.Items)
	}

	response.WriteAsJson(result)
}

func (h *handler) GetKubeconfig(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	kubectlConfig, err := h.kubeconfigOperator.GetKubeConfig(request.Request.Context(), username)

	if err != nil {
		klog.Error(err)
		if k8serr.IsNotFound(err) {
			// recreate
			response.WriteHeaderAndJson(http.StatusNotFound, errors.Wrap(err), restful.MIME_JSON)
		} else {
			response.WriteHeaderAndJson(http.StatusInternalServerError, errors.Wrap(err), restful.MIME_JSON)
		}
		return
	}

	response.Write([]byte(kubectlConfig))
}
