/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"context"

	k8suitl "kubesphere.io/kubesphere/pkg/utils/k8sutil"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	appv2 "kubesphere.io/api/application/v2"

	"kubesphere.io/kubesphere/pkg/api"
)

func (h *appHandler) AppCrList(req *restful.Request, resp *restful.Response) {
	clusterName := req.QueryParameter("cluster")
	gvr := schema.GroupVersionResource{
		Group:    req.QueryParameter("group"),
		Version:  req.QueryParameter("version"),
		Resource: req.QueryParameter("resource"),
	}
	opts := metav1.ListOptions{}
	labelSelectorStr := req.QueryParameter("labelSelector")
	labelSelector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	opts.LabelSelector = labelSelector.String()
	_, dynamicClient, _, err := h.getCluster(req, clusterName)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	list, err := dynamicClient.Resource(gvr).List(context.Background(), opts)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(k8suitl.ConvertToListResult(list, req))
}

func (h *appHandler) CreateOrUpdateCR(req *restful.Request, resp *restful.Response) {
	gvr := schema.GroupVersionResource{
		Group:    req.QueryParameter("group"),
		Version:  req.QueryParameter("version"),
		Resource: req.QueryParameter("resource"),
	}
	obj := unstructured.Unstructured{}
	err := req.ReadEntity(&obj)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	lbs := obj.GetLabels()
	if lbs == nil {
		lbs = make(map[string]string)
	}
	lbs[appv2.AppReleaseReferenceLabelKey] = req.PathParameter("application")
	obj.SetLabels(lbs)

	js, err := obj.MarshalJSON()
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	clusterName := req.QueryParameter("cluster")
	_, dynamicClient, _, err := h.getCluster(req, clusterName)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	opt := metav1.PatchOptions{FieldManager: "applicationOperator"}
	_, err = dynamicClient.Resource(gvr).
		Namespace(obj.GetNamespace()).
		Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, js, opt)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
}

func (h *appHandler) DescribeAppCr(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("crname")
	namespace := req.QueryParameter("namespace")

	gvr := schema.GroupVersionResource{
		Group:    req.QueryParameter("group"),
		Version:  req.QueryParameter("version"),
		Resource: req.QueryParameter("resource"),
	}
	clusterName := req.QueryParameter("cluster")
	_, dynamicClient, _, err := h.getCluster(req, clusterName)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	get, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	get.SetManagedFields(nil)
	resp.WriteEntity(get)
}

func (h *appHandler) DeleteAppCr(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("crname")
	namespace := req.QueryParameter("namespace")
	gvr := schema.GroupVersionResource{
		Group:    req.QueryParameter("group"),
		Version:  req.QueryParameter("version"),
		Resource: req.QueryParameter("resource"),
	}
	clusterName := req.QueryParameter("cluster")
	_, dynamicClient, _, err := h.getCluster(req, clusterName)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	err = dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
}
