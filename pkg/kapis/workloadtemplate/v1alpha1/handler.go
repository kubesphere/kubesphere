package v1alpha1

import (
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	k8suitl "kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

func (h *templateHandler) list(req *restful.Request, resp *restful.Response) {
	cmList := corev1.ConfigMapList{}
	requirements, _ := labels.SelectorFromSet(map[string]string{SchemeGroupVersion.Group: "true"}).Requirements()
	userSelector := query.ParseQueryParameter(req).Selector()
	combinedSelector := labels.NewSelector()
	combinedSelector = combinedSelector.Add(requirements...)
	if userSelector != nil {
		userRequirements, _ := userSelector.Requirements()
		combinedSelector = combinedSelector.Add(userRequirements...)
	}

	opts := []client.ListOption{
		client.InNamespace(constants.KubeSphereNamespace),
		client.MatchingLabelsSelector{Selector: combinedSelector},
	}

	err := h.client.List(req.Request.Context(), &cmList, opts...)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	workspace := req.PathParameter("workspace")
	filteredList := &corev1.ConfigMapList{}
	for _, cm := range cmList.Items {
		if workspace != "" {
			if !stringutils.StringIn(cm.Labels[constants.WorkspaceLabelKey], []string{workspace}) {
				continue
			}
		}
		filteredList.Items = append(filteredList.Items, cm)
	}

	resp.WriteEntity(k8suitl.ConvertToListResult(filteredList, req))
}

func (h *templateHandler) apply(req *restful.Request, resp *restful.Response) {
	cm := &corev1.ConfigMap{}
	err := req.ReadEntity(cm)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	newCm := &corev1.ConfigMap{}
	newCm.Name = cm.Name
	newCm.Namespace = constants.KubeSphereNamespace
	mutateFn := func() error {
		newCm.Annotations = cm.Annotations
		newCm.Labels = cm.Labels
		if newCm.Labels == nil {
			newCm.Labels = make(map[string]string)
		}
		newCm.Labels[SchemeGroupVersion.Group] = "true"
		workspace := req.PathParameter("workspace")
		if workspace != "" {
			newCm.Labels[constants.WorkspaceLabelKey] = workspace
		}
		newCm.Data = cm.Data
		return nil
	}
	_, err = controllerutil.CreateOrUpdate(req.Request.Context(), h.client, newCm, mutateFn)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	newCm.SetManagedFields(nil)
	resp.WriteAsJson(newCm)
}

func (h *templateHandler) delete(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("workloadtemplate")
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name}}
	cm.Namespace = constants.KubeSphereNamespace
	err := h.client.Delete(req.Request.Context(), cm)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(errors.None)
}

func (h *templateHandler) get(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("workloadtemplate")
	cm := &corev1.ConfigMap{}
	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: name, Namespace: constants.KubeSphereNamespace}, cm)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	cm.SetManagedFields(nil)
	resp.WriteAsJson(cm)
}
