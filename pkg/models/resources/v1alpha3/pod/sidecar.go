/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package pod

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ContainerType classifies a container's role within a pod.
type ContainerType string

const (
	// ContainerTypeMain is a regular application container.
	ContainerTypeMain ContainerType = "main"
	// ContainerTypeSidecar is a sidecar container — either a native K8s sidecar
	// (an init container with restartPolicy=Always, K8s 1.28+) or a well-known
	// mesh/logging/monitoring sidecar identified by naming convention.
	ContainerTypeSidecar ContainerType = "sidecar"
	// ContainerTypeInit is a traditional init container (runs to completion before
	// app containers start; does NOT have restartPolicy=Always).
	ContainerTypeInit ContainerType = "init"
)

// wellKnownSidecarNames contains lower-case substrings that identify commonly
// deployed sidecar containers (service mesh proxies, log shippers, tracing
// agents, etc.).  A container whose name contains any of these substrings is
// classified as ContainerTypeSidecar.
var wellKnownSidecarNames = []string{
	"istio-proxy",
	"istio-init",
	"envoy",
	"linkerd-proxy",
	"linkerd-init",
	"jaeger-agent",
	"zipkin",
	"datadog-agent",
	"fluentd",
	"fluent-bit",
	"filebeat",
	"logstash",
	"promtail",
	"otel-collector",
	"opentelemetry-collector",
	"vault-agent",
}

// ContainerInfo holds the classified metadata for a single container inside a pod.
type ContainerInfo struct {
	// Name is the container name.
	Name string `json:"name"`
	// Image is the container image reference.
	Image string `json:"image"`
	// Type is one of "main", "sidecar", or "init".
	Type ContainerType `json:"type"`
	// Ready indicates whether the container is currently ready.
	Ready bool `json:"ready"`
	// RestartCount is the number of times the container has been restarted.
	RestartCount int32 `json:"restartCount"`
	// State is a human-readable representation of the container's current state
	// (Running, Waiting, Terminated, or Unknown).
	State string `json:"state"`
	// StateDetail provides the reason or message for the current state, if available.
	StateDetail string `json:"stateDetail,omitempty"`
}

// ClassifyPodContainers returns a ContainerInfo slice that covers every container
// in the pod (init containers, native sidecar init containers, and regular
// containers), each labelled with its ContainerType.
func ClassifyPodContainers(pod *corev1.Pod) []ContainerInfo {
	// Build fast-lookup maps from container name → status.
	initStatusByName := make(map[string]corev1.ContainerStatus, len(pod.Status.InitContainerStatuses))
	for _, s := range pod.Status.InitContainerStatuses {
		initStatusByName[s.Name] = s
	}
	statusByName := make(map[string]corev1.ContainerStatus, len(pod.Status.ContainerStatuses))
	for _, s := range pod.Status.ContainerStatuses {
		statusByName[s.Name] = s
	}

	var result []ContainerInfo

	// Process init containers first (preserves declaration order).
	for _, c := range pod.Spec.InitContainers {
		st := initStatusByName[c.Name]
		ctype := ContainerTypeInit
		if isNativeSidecar(c) {
			// Native K8s sidecar: init container with restartPolicy=Always.
			ctype = ContainerTypeSidecar
		}
		state, detail := describeContainerState(st.State)
		result = append(result, ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			Type:         ctype,
			Ready:        st.Ready,
			RestartCount: st.RestartCount,
			State:        state,
			StateDetail:  detail,
		})
	}

	// Process regular containers.
	for _, c := range pod.Spec.Containers {
		st := statusByName[c.Name]
		ctype := ContainerTypeMain
		if isWellKnownSidecar(c.Name) {
			ctype = ContainerTypeSidecar
		}
		state, detail := describeContainerState(st.State)
		result = append(result, ContainerInfo{
			Name:         c.Name,
			Image:        c.Image,
			Type:         ctype,
			Ready:        st.Ready,
			RestartCount: st.RestartCount,
			State:        state,
			StateDetail:  detail,
		})
	}

	return result
}

// isNativeSidecar returns true if c is a native K8s sidecar — i.e. an init
// container whose RestartPolicy is explicitly set to "Always" (SidecarContainers
// feature gate, K8s 1.28+).
func isNativeSidecar(c corev1.Container) bool {
	return c.RestartPolicy != nil &&
		*c.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

// isWellKnownSidecar returns true when the container name matches a known
// sidecar naming pattern.  The match is case-insensitive substring match.
func isWellKnownSidecar(name string) bool {
	lower := strings.ToLower(name)
	for _, pattern := range wellKnownSidecarNames {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// describeContainerState converts a ContainerState to a human-readable string
// pair: (state, detail).
func describeContainerState(state corev1.ContainerState) (string, string) {
	switch {
	case state.Running != nil:
		return "Running", ""
	case state.Waiting != nil:
		detail := state.Waiting.Message
		if detail == "" {
			detail = state.Waiting.Reason
		}
		return "Waiting", detail
	case state.Terminated != nil:
		detail := state.Terminated.Message
		if detail == "" {
			detail = state.Terminated.Reason
		}
		return "Terminated", detail
	default:
		return "Unknown", ""
	}
}
