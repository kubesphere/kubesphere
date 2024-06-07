/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/api"
	apiv1alpha1 "kubesphere.io/kubesphere/pkg/api/cluster/v1alpha1"
)

func labelExists(req apiv1alpha1.CreateLabelRequest, labels *clusterv1alpha1.LabelList) bool {
	for _, label := range labels.Items {
		if label.Spec.Key == req.Key && label.Spec.Value == req.Value {
			return true
		}
	}
	return false
}

func (h *handler) createLabels(request *restful.Request, response *restful.Response) {
	var labelRequests []apiv1alpha1.CreateLabelRequest
	if err := request.ReadEntity(&labelRequests); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	allLabels := &clusterv1alpha1.LabelList{}
	if err := h.client.List(context.Background(), allLabels); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	results := make([]*clusterv1alpha1.Label, 0)
	for _, r := range labelRequests {
		if labelExists(r, allLabels) {
			api.HandleBadRequest(response, request, fmt.Errorf("label %s/%s already exists", r.Key, r.Value))
			return
		}

		obj := &clusterv1alpha1.Label{
			ObjectMeta: metav1.ObjectMeta{
				Name:       rand.String(6),
				Finalizers: []string{clusterv1alpha1.LabelFinalizer},
			},
			Spec: clusterv1alpha1.LabelSpec{
				Key:   strings.TrimSpace(r.Key),
				Value: strings.TrimSpace(r.Value),
			},
		}
		if err := h.client.Create(context.Background(), obj); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		results = append(results, obj)
	}
	response.WriteEntity(results)
}

func (h *handler) updateLabel(request *restful.Request, response *restful.Response) {
	label := &clusterv1alpha1.Label{}
	if err := h.client.Get(context.Background(), types.NamespacedName{Name: request.PathParameter("label")}, label); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	switch request.QueryParameter("action") {
	case "unbind": // unbind clusters
		var unbindRequest apiv1alpha1.UnbindClustersRequest
		if err := request.ReadEntity(&unbindRequest); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		for _, name := range unbindRequest.Clusters {
			cluster := &clusterv1alpha1.Cluster{}
			if err := h.client.Get(context.Background(), types.NamespacedName{Name: name}, cluster); err != nil {
				api.HandleBadRequest(response, request, err)
				return
			}
			cluster = cluster.DeepCopy()
			delete(cluster.Labels, fmt.Sprintf(clusterv1alpha1.ClusterLabelFormat, label.Name))
			if err := h.client.Update(context.Background(), cluster); err != nil {
				api.HandleBadRequest(response, request, err)
				return
			}
		}
		clusters := sets.NewString(label.Spec.Clusters...)
		clusters.Delete(unbindRequest.Clusters...)
		label.Spec.Clusters = clusters.List()
		if err := h.client.Update(context.Background(), label); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		response.WriteEntity(label)
	default: // update label key/value
		var labelRequest apiv1alpha1.CreateLabelRequest
		if err := request.ReadEntity(&labelRequest); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		allLabels := &clusterv1alpha1.LabelList{}
		if err := h.client.List(context.Background(), allLabels); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		if labelExists(labelRequest, allLabels) {
			api.HandleBadRequest(response, request, fmt.Errorf("label %s/%s already exists", labelRequest.Key, labelRequest.Value))
			return
		}
		label.Spec.Key = strings.TrimSpace(labelRequest.Key)
		label.Spec.Value = strings.TrimSpace(labelRequest.Value)
		if err := h.client.Update(context.Background(), label); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		response.WriteEntity(label)
	}
}

func (h *handler) deleteLabels(request *restful.Request, response *restful.Response) {
	var names []string
	if err := request.ReadEntity(&names); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, name := range names {
		label := &clusterv1alpha1.Label{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		if err := h.client.Delete(context.Background(), label); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
	}
	response.WriteHeader(http.StatusOK)
}

func (h *handler) bindingClusters(request *restful.Request, response *restful.Response) {
	var bindingRequest apiv1alpha1.BindingClustersRequest
	if err := request.ReadEntity(&bindingRequest); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, name := range bindingRequest.Labels {
		label := &clusterv1alpha1.Label{}
		if err := h.client.Get(context.Background(), types.NamespacedName{Name: name}, label); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		label.Spec.Clusters = append(label.Spec.Clusters, bindingRequest.Clusters...)
		if err := h.client.Update(context.Background(), label); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
	}

	response.WriteHeader(http.StatusOK)
}

func (h *handler) listLabelGroups(request *restful.Request, response *restful.Response) {
	allLabels := &clusterv1alpha1.LabelList{}
	if err := h.client.List(context.Background(), allLabels); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	results := make(map[string][]apiv1alpha1.LabelValue)
	for _, label := range allLabels.Items {
		results[label.Spec.Key] = append(results[label.Spec.Key], apiv1alpha1.LabelValue{
			Value: label.Spec.Value,
			ID:    label.Name,
		})
	}

	response.WriteEntity(results)
}
