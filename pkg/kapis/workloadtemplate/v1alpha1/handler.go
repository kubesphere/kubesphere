/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apiserver/pkg/authentication/user"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"

	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/server/errors"
	k8suitl "kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

func (h *templateHandler) listWorkloadTemplate(req *restful.Request, resp *restful.Response) {
	secretList := corev1.SecretList{}
	requirements, _ := labels.SelectorFromSet(map[string]string{SchemeGroupVersion.Group: "true"}).Requirements()
	userSelector := query.ParseQueryParameter(req).Selector()
	combinedSelector := labels.NewSelector()
	combinedSelector = combinedSelector.Add(requirements...)
	if userSelector != nil {
		userRequirements, _ := userSelector.Requirements()
		combinedSelector = combinedSelector.Add(userRequirements...)
	}
	opts := []client.ListOption{
		client.MatchingLabelsSelector{Selector: combinedSelector},
	}
	namespace := req.PathParameter("namespace")
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	err := h.client.List(req.Request.Context(), &secretList, opts...)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	workspace := req.PathParameter("workspace")
	if workspace == "" {
		resp.WriteEntity(k8suitl.ConvertToListResult(&secretList, req))
		return
	}

	user, ok := request.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	filteredList, err := h.FilterByPermissions(workspace, user, secretList)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(k8suitl.ConvertToListResult(filteredList, req))
}

func (h *templateHandler) FilterByPermissions(workspace string, user user.Info, secretList corev1.SecretList) (*corev1.SecretList, error) {

	listNS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Workspace:       workspace,
		Resource:        "namespaces",
		ResourceRequest: true,
		ResourceScope:   request.WorkspaceScope,
	}

	decision, _, err := h.authorizer.Authorize(listNS)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var namespaceList []string
	if decision == authorizer.DecisionAllow {
		queryParam := query.New()
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1beta1.WorkspaceLabel, workspace))
		result, err := h.resourceGetter.List("namespaces", "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		for _, item := range result.Items {
			ns := item.(*corev1.Namespace)
			listWorkLoadTemplate := authorizer.AttributesRecord{
				User:            user,
				Verb:            "list",
				Namespace:       ns.Name,
				Resource:        "workloadtemplates",
				ResourceRequest: true,
				ResourceScope:   request.NamespaceScope,
			}
			decision, _, err = h.authorizer.Authorize(listWorkLoadTemplate)
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			if decision == authorizer.DecisionAllow {
				namespaceList = append(namespaceList, ns.Name)
			} else {
				klog.Infof("user %s has no permission to list workloadtemplate in namespace %s", user.GetName(), ns.Name)
			}
		}
	}

	filteredList := &corev1.SecretList{}
	for _, item := range secretList.Items {
		if !stringutils.StringIn(item.Namespace, namespaceList) {
			continue
		}
		filteredList.Items = append(filteredList.Items, item)
	}
	return filteredList, nil
}

func (h *templateHandler) applyWorkloadTemplate(req *restful.Request, resp *restful.Response) {
	secret := &corev1.Secret{}
	err := req.ReadEntity(secret)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if req.PathParameter("workloadtemplate") == "" {
		// create new
		ns := req.PathParameter("namespace")
		err = h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: secret.Name, Namespace: ns}, secret)
		if err != nil && !apierrors.IsNotFound(err) {
			api.HandleError(resp, req, err)
			return
		}
		if err == nil {
			api.HandleConflict(resp, req, fmt.Errorf("workloadtemplate %s already exists", secret.Name))
			return
		}
	}

	newSecret := &corev1.Secret{}
	newSecret.Name = secret.Name
	newSecret.Namespace = req.PathParameter("namespace")
	mutateFn := func() error {
		newSecret.Annotations = secret.Annotations
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		newSecret.Labels = secret.Labels
		newSecret.Labels[SchemeGroupVersion.Group] = "true"
		newSecret.StringData = secret.StringData
		newSecret.Type = corev1.SecretType(fmt.Sprintf("%s/%s", SchemeGroupVersion.Group, "workloadtemplate"))
		return nil
	}
	_, err = controllerutil.CreateOrUpdate(req.Request.Context(), h.client, newSecret, mutateFn)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	newSecret.SetManagedFields(nil)
	resp.WriteAsJson(newSecret)
}

func (h *templateHandler) deleteWorkloadTemplate(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("workloadtemplate")
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name}}
	secret.Namespace = req.PathParameter("namespace")
	err := h.client.Delete(req.Request.Context(), secret)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(errors.None)
}

func (h *templateHandler) getWorkloadTemplate(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("workloadtemplate")
	secret := &corev1.Secret{}
	ns := req.PathParameter("namespace")
	err := h.client.Get(req.Request.Context(), runtimeclient.ObjectKey{Name: name, Namespace: ns}, secret)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	secret.SetManagedFields(nil)
	resp.WriteAsJson(secret)
}
