/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "workloadnetwork.kubesphere.io", Version: "v1alpha1"}
)

type networkHandler struct {
	client runtimeclient.Client
}

// NewHandler creates a new network configuration handler
func NewHandler(client runtimeclient.Client) rest.Handler {
	return &networkHandler{
		client: client,
	}
}

// AddToContainer adds the network configuration API routes to the container
func (h *networkHandler) AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(SchemeGroupVersion)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{kind}/{name}/networkconfig").
		To(h.getWorkloadNetworkConfig).
		Doc("Get network configuration for a workload").
		Notes("Retrieves the network configuration (DNS, host network, etc.) for a specific workload.").
		Operation("getWorkloadNetworkConfig").
		Param(ws.PathParameter("namespace", "Namespace of the workload").Required(true)).
		Param(ws.PathParameter("kind", "Kind of workload (Deployment, StatefulSet, DaemonSet, Job)").Required(true)).
		Param(ws.PathParameter("name", "Name of the workload").Required(true)).
		Returns(http.StatusOK, api.StatusOK, WorkloadNetworkConfigResponse{}).
		Metadata("tag", api.TagNamespacedResources))

	ws.Route(ws.PUT("/namespaces/{namespace}/workloads/{kind}/{name}/networkconfig").
		To(h.updateWorkloadNetworkConfig).
		Doc("Update network configuration for a workload").
		Notes("Updates the network configuration (DNS, host network, etc.) for a specific workload.").
		Operation("updateWorkloadNetworkConfig").
		Param(ws.PathParameter("namespace", "Namespace of the workload").Required(true)).
		Param(ws.PathParameter("kind", "Kind of workload (Deployment, StatefulSet, DaemonSet, Job)").Required(true)).
		Param(ws.PathParameter("name", "Name of the workload").Required(true)).
		Reads(WorkloadNetworkConfigRequest{}).
		Returns(http.StatusOK, api.StatusOK, WorkloadNetworkConfigResponse{}).
		Metadata("tag", api.TagNamespacedResources))

	c.Add(ws)
	return nil
}
