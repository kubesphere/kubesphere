// Copyright (c) 2016-2021 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conversion

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	kapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/lib/numorstring"

	libapiv3 "github.com/projectcalico/calico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/calico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/calico/libcalico-go/lib/json"
	"github.com/projectcalico/calico/libcalico-go/lib/names"
	cnet "github.com/projectcalico/calico/libcalico-go/lib/net"
)

type defaultWorkloadEndpointConverter struct{}

// VethNameForWorkload returns a deterministic veth name
// for the given Kubernetes workload (WEP) name and namespace.
func (wc defaultWorkloadEndpointConverter) VethNameForWorkload(namespace, podname string) string {
	// A SHA1 is always 20 bytes long, and so is sufficient for generating the
	// veth name and mac addr.
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s.%s", namespace, podname)))
	prefix := os.Getenv("FELIX_INTERFACEPREFIX")
	if prefix == "" {
		// Prefix is not set. Default to "cali"
		prefix = "cali"
	} else {
		// Prefix is set - use the first value in the list.
		splits := strings.Split(prefix, ",")
		prefix = splits[0]
	}
	log.WithField("prefix", prefix).Debugf("Using prefix to create a WorkloadEndpoint veth name")
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(h.Sum(nil))[:11])
}

func (wc defaultWorkloadEndpointConverter) PodToWorkloadEndpoints(pod *kapiv1.Pod) ([]*model.KVPair, error) {
	wep, err := wc.podToDefaultWorkloadEndpoint(pod)
	if err != nil {
		return nil, err
	}

	return []*model.KVPair{wep}, nil
}

// PodToWorkloadEndpoint converts a Pod to a WorkloadEndpoint.  It assumes the calling code
// has verified that the provided Pod is valid to convert to a WorkloadEndpoint.
// PodToWorkloadEndpoint requires a Pods Name and Node Name to be populated. It will
// fail to convert from a Pod to WorkloadEndpoint otherwise.
func (wc defaultWorkloadEndpointConverter) podToDefaultWorkloadEndpoint(pod *kapiv1.Pod) (*model.KVPair, error) {
	log.WithField("pod", pod).Debug("Converting pod to WorkloadEndpoint")
	// Get all the profiles that apply
	var profiles []string

	// Pull out the Namespace based profile off the pod name and Namespace.
	profiles = append(profiles, NamespaceProfileNamePrefix+pod.Namespace)

	// Pull out the Serviceaccount based profile off the pod SA and namespace
	if pod.Spec.ServiceAccountName != "" {
		profiles = append(profiles, serviceAccountNameToProfileName(pod.Spec.ServiceAccountName, pod.Namespace))
	}

	wepids := names.WorkloadEndpointIdentifiers{
		Node:         pod.Spec.NodeName,
		Orchestrator: apiv3.OrchestratorKubernetes,
		Endpoint:     "eth0",
		Pod:          pod.Name,
	}
	wepName, err := wepids.CalculateWorkloadEndpointName(false)
	if err != nil {
		return nil, err
	}

	podIPNets, err := getPodIPs(pod)
	if err != nil {
		// IP address was present but malformed in some way, handle as an explicit failure.
		return nil, err
	}

	if IsFinished(pod) {
		// Pod is finished but not yet deleted.  In this state the IP will have been freed and returned to the pool
		// so we need to make sure we don't let the caller believe it still belongs to this endpoint.
		// Pods with no IPs will get filtered out before they get to Felix in the watcher syncer cache layer.
		// We can't pretend the workload endpoint is deleted _here_ because that would confuse users of the
		// native v3 Watch() API.
		log.Debug("Pod is in a 'finished' state so no longer owns its IP(s).")
		podIPNets = nil
	}

	ipNets := []string{}
	for _, ipNet := range podIPNets {
		ipNets = append(ipNets, ipNet.String())
	}

	// Generate the interface name based on workload.  This must match
	// the host-side veth configured by the CNI plugin.
	interfaceName := wc.VethNameForWorkload(pod.Namespace, pod.Name)

	// Build the labels map.  Start with the pod labels, and append two additional labels for
	// namespace and orchestrator matches.
	labels := pod.Labels
	if labels == nil {
		labels = make(map[string]string, 2)
	}
	labels[apiv3.LabelNamespace] = pod.Namespace
	labels[apiv3.LabelOrchestrator] = apiv3.OrchestratorKubernetes

	if pod.Spec.ServiceAccountName != "" && len(pod.Spec.ServiceAccountName) < 63 {
		// For backwards compatibility, include the label if less than 63 characters.
		labels[apiv3.LabelServiceAccount] = pod.Spec.ServiceAccountName
	}

	// Pull out floating IP annotation
	var floatingIPs []libapiv3.IPNAT
	if annotation, ok := pod.Annotations["cni.projectcalico.org/floatingIPs"]; ok && len(podIPNets) > 0 {

		// Parse Annotation data
		var ips []string
		err := json.Unmarshal([]byte(annotation), &ips)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s' as JSON: %s", annotation, err)
		}

		// Get IPv4 and IPv6 targets for NAT
		var podnetV4, podnetV6 *cnet.IPNet
		for _, ipNet := range podIPNets {
			if ipNet.IP.To4() != nil {
				podnetV4 = ipNet
				netmask, _ := podnetV4.Mask.Size()
				if netmask != 32 {
					return nil, fmt.Errorf("PodIP %v is not a valid IPv4: Mask size is %d, not 32", ipNet, netmask)
				}
			} else {
				podnetV6 = ipNet
				netmask, _ := podnetV6.Mask.Size()
				if netmask != 128 {
					return nil, fmt.Errorf("PodIP %v is not a valid IPv6: Mask size is %d, not 128", ipNet, netmask)
				}
			}
		}

		for _, ip := range ips {
			if strings.Contains(ip, ":") {
				if podnetV6 != nil {
					floatingIPs = append(floatingIPs, libapiv3.IPNAT{
						InternalIP: podnetV6.IP.String(),
						ExternalIP: ip,
					})
				}
			} else {
				if podnetV4 != nil {
					floatingIPs = append(floatingIPs, libapiv3.IPNAT{
						InternalIP: podnetV4.IP.String(),
						ExternalIP: ip,
					})
				}
			}
		}
	}

	// Handle source IP spoofing annotation
	var sourcePrefixes []string
	if annotation, ok := pod.Annotations["cni.projectcalico.org/allowedSourcePrefixes"]; ok && annotation != "" {
		// Parse Annotation data
		var requestedSourcePrefixes []string
		err := json.Unmarshal([]byte(annotation), &requestedSourcePrefixes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s' as JSON: %s", annotation, err)
		}

		// Filter out any invalid entries and normalize the CIDRs.
		for _, prefix := range requestedSourcePrefixes {
			if _, n, err := cnet.ParseCIDR(prefix); err != nil {
				return nil, fmt.Errorf("failed to parse '%s' as a CIDR: %s", prefix, err)
			} else {
				sourcePrefixes = append(sourcePrefixes, n.String())
			}
		}

	}

	// Map any named ports through.
	var endpointPorts []libapiv3.WorkloadEndpointPort
	for _, container := range pod.Spec.Containers {
		for _, containerPort := range container.Ports {
			if containerPort.ContainerPort != 0 && (containerPort.HostPort != 0 || containerPort.Name != "") {
				var modelProto numorstring.Protocol
				switch containerPort.Protocol {
				case kapiv1.ProtocolUDP:
					modelProto = numorstring.ProtocolFromString("udp")
				case kapiv1.ProtocolSCTP:
					modelProto = numorstring.ProtocolFromString("sctp")
				case kapiv1.ProtocolTCP, kapiv1.Protocol("") /* K8s default is TCP. */ :
					modelProto = numorstring.ProtocolFromString("tcp")
				default:
					log.WithFields(log.Fields{
						"protocol": containerPort.Protocol,
						"pod":      pod,
						"port":     containerPort,
					}).Debug("Ignoring named port with unknown protocol")
					continue
				}

				endpointPorts = append(endpointPorts, libapiv3.WorkloadEndpointPort{
					Name:     containerPort.Name,
					Protocol: modelProto,
					Port:     uint16(containerPort.ContainerPort),
					HostPort: uint16(containerPort.HostPort),
					HostIP:   containerPort.HostIP,
				})
			}
		}
	}

	// Get the container ID if present.  This is used in the CNI plugin to distinguish different pods that have
	// the same name.  For example, restarted stateful set pods.
	containerID := pod.Annotations[AnnotationContainerID]

	// Create the workload endpoint.
	wep := libapiv3.NewWorkloadEndpoint()
	wep.ObjectMeta = metav1.ObjectMeta{
		Name:              wepName,
		Namespace:         pod.Namespace,
		CreationTimestamp: pod.CreationTimestamp,
		UID:               pod.UID,
		Labels:            labels,
		GenerateName:      pod.GenerateName,
	}
	wep.Spec = libapiv3.WorkloadEndpointSpec{
		Orchestrator:               "k8s",
		Node:                       pod.Spec.NodeName,
		Pod:                        pod.Name,
		ContainerID:                containerID,
		Endpoint:                   "eth0",
		InterfaceName:              interfaceName,
		Profiles:                   profiles,
		IPNetworks:                 ipNets,
		Ports:                      endpointPorts,
		IPNATs:                     floatingIPs,
		ServiceAccountName:         pod.Spec.ServiceAccountName,
		AllowSpoofedSourcePrefixes: sourcePrefixes,
	}

	if v, ok := pod.Annotations["k8s.v1.cni.cncf.io/network-status"]; ok {
		if wep.Annotations == nil {
			wep.Annotations = make(map[string]string)
		}
		wep.Annotations["k8s.v1.cni.cncf.io/network-status"] = v
	}

	// Embed the workload endpoint into a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name:      wepName,
			Namespace: pod.Namespace,
			Kind:      libapiv3.KindWorkloadEndpoint,
		},
		Value:    wep,
		Revision: pod.ResourceVersion,
	}
	return &kvp, nil
}
