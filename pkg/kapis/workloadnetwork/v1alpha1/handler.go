/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
)

const (
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindJob         = "Job"
)

// getWorkloadNetworkConfig handles GET requests for workload network configuration
func (h *networkHandler) getWorkloadNetworkConfig(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")
	kind := req.PathParameter("kind")

	networkConfig, err := h.getNetworkConfigFromWorkload(req, namespace, name, kind)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}

	response := &WorkloadNetworkConfigResponse{
		Kind:          kind,
		Name:          name,
		Namespace:     namespace,
		NetworkConfig: *networkConfig,
	}

	resp.WriteAsJson(response)
}

// updateWorkloadNetworkConfig handles PUT requests to update workload network configuration
func (h *networkHandler) updateWorkloadNetworkConfig(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")
	kind := req.PathParameter("kind")

	request := &WorkloadNetworkConfigRequest{}
	if err := req.ReadEntity(request); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if request.Kind != "" && request.Kind != kind {
		api.HandleBadRequest(resp, req, fmt.Errorf("kind in request body (%s) does not match path parameter (%s)", request.Kind, kind))
		return
	}

	err := h.updateNetworkConfig(req, namespace, name, kind, &request.NetworkConfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}

	updatedConfig, err := h.getNetworkConfigFromWorkload(req, namespace, name, kind)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	response := &WorkloadNetworkConfigResponse{
		Kind:          kind,
		Name:          name,
		Namespace:     namespace,
		NetworkConfig: *updatedConfig,
	}

	resp.WriteAsJson(response)
}

// getNetworkConfigFromWorkload extracts the network configuration from a workload
func (h *networkHandler) getNetworkConfigFromWorkload(req *restful.Request, namespace, name, kind string) (*NetworkConfig, error) {
	ctx := req.Request.Context()
	key := runtimeclient.ObjectKey{Namespace: namespace, Name: name}

	var podSpec *corev1.PodSpec
	switch kind {
	case KindDeployment:
		deployment := &appsv1.Deployment{}
		if err := h.client.Get(ctx, key, deployment); err != nil {
			return nil, err
		}
		podSpec = &deployment.Spec.Template.Spec
	case KindStatefulSet:
		statefulSet := &appsv1.StatefulSet{}
		if err := h.client.Get(ctx, key, statefulSet); err != nil {
			return nil, err
		}
		podSpec = &statefulSet.Spec.Template.Spec
	case KindDaemonSet:
		daemonSet := &appsv1.DaemonSet{}
		if err := h.client.Get(ctx, key, daemonSet); err != nil {
			return nil, err
		}
		podSpec = &daemonSet.Spec.Template.Spec
	case KindJob:
		job := &batchv1.Job{}
		if err := h.client.Get(ctx, key, job); err != nil {
			return nil, err
		}
		podSpec = &job.Spec.Template.Spec
	default:
		return nil, fmt.Errorf("unsupported workload kind: %s", kind)
	}

	return extractNetworkConfig(podSpec), nil
}

// updateNetworkConfig updates the network configuration on a workload
func (h *networkHandler) updateNetworkConfig(req *restful.Request, namespace, name, kind string, config *NetworkConfig) error {
	ctx := req.Request.Context()
	key := runtimeclient.ObjectKey{Namespace: namespace, Name: name}

	switch kind {
	case KindDeployment:
		deployment := &appsv1.Deployment{}
		if err := h.client.Get(ctx, key, deployment); err != nil {
			return err
		}
		applyNetworkConfig(&deployment.Spec.Template.Spec, config)
		return h.client.Update(ctx, deployment)
	case KindStatefulSet:
		statefulSet := &appsv1.StatefulSet{}
		if err := h.client.Get(ctx, key, statefulSet); err != nil {
			return err
		}
		applyNetworkConfig(&statefulSet.Spec.Template.Spec, config)
		return h.client.Update(ctx, statefulSet)
	case KindDaemonSet:
		daemonSet := &appsv1.DaemonSet{}
		if err := h.client.Get(ctx, key, daemonSet); err != nil {
			return err
		}
		applyNetworkConfig(&daemonSet.Spec.Template.Spec, config)
		return h.client.Update(ctx, daemonSet)
	case KindJob:
		job := &batchv1.Job{}
		if err := h.client.Get(ctx, key, job); err != nil {
			return err
		}
		applyNetworkConfig(&job.Spec.Template.Spec, config)
		return h.client.Update(ctx, job)
	default:
		return fmt.Errorf("unsupported workload kind: %s", kind)
	}
}

// extractNetworkConfig extracts network configuration from a PodSpec
func extractNetworkConfig(podSpec *corev1.PodSpec) *NetworkConfig {
	return &NetworkConfig{
		DNSPolicy:             podSpec.DNSPolicy,
		DNSConfig:             podSpec.DNSConfig,
		HostNetwork:           podSpec.HostNetwork,
		HostPID:               podSpec.HostPID,
		HostIPC:               podSpec.HostIPC,
		Hostname:              podSpec.Hostname,
		Subdomain:             podSpec.Subdomain,
		HostAliases:           podSpec.HostAliases,
		ShareProcessNamespace: podSpec.ShareProcessNamespace,
	}
}

// applyNetworkConfig applies network configuration to a PodSpec
func applyNetworkConfig(podSpec *corev1.PodSpec, config *NetworkConfig) {
	if config.DNSPolicy != "" {
		podSpec.DNSPolicy = config.DNSPolicy
	}
	if config.DNSConfig != nil {
		podSpec.DNSConfig = config.DNSConfig
	}
	podSpec.HostNetwork = config.HostNetwork
	podSpec.HostPID = config.HostPID
	podSpec.HostIPC = config.HostIPC
	if config.Hostname != "" {
		podSpec.Hostname = config.Hostname
	}
	if config.Subdomain != "" {
		podSpec.Subdomain = config.Subdomain
	}
	if config.HostAliases != nil {
		podSpec.HostAliases = config.HostAliases
	}
	if config.ShareProcessNamespace != nil {
		podSpec.ShareProcessNamespace = config.ShareProcessNamespace
	}
}
