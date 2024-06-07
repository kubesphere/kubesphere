/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	appv2 "kubesphere.io/api/application/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	_, dynamicClient, _, err := h.getCluster(clusterName)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	list, err := dynamicClient.Resource(gvr).List(context.Background(), opts)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(convertToListResult(list, req))
}
func checkPermissions(gvr schema.GroupVersionResource, app appv2.Application) (allow bool) {
	for _, i := range app.Spec.Resources {
		if gvr.Resource == i.Resource {
			allow = true
			break
		}
	}
	return allow
}
func (h *appHandler) CreateOrUpdateCR(req *restful.Request, resp *restful.Response) {
	gvr := schema.GroupVersionResource{
		Group:    req.QueryParameter("group"),
		Version:  req.QueryParameter("version"),
		Resource: req.QueryParameter("resource"),
	}
	appID := req.QueryParameter("app")
	app := appv2.Application{}
	err := h.client.Get(req.Request.Context(), client.ObjectKey{Name: appID}, &app)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	allow := checkPermissions(gvr, app)
	if !allow {
		api.HandleForbidden(resp, nil, fmt.Errorf("resource %s not allow", gvr.Resource))
		return
	}
	obj := unstructured.Unstructured{}
	err = req.ReadEntity(&obj)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	js, err := obj.MarshalJSON()
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	clusterName := req.QueryParameter("cluster")
	_, dynamicClient, _, err := h.getCluster(clusterName)
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
	_, dynamicClient, _, err := h.getCluster(clusterName)
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
	_, dynamicClient, _, err := h.getCluster(clusterName)
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
