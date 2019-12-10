// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.

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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	kapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/libcalico-go/lib/names"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/numorstring"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	protoTCP = kapiv1.ProtocolTCP
)

type selectorType int8

const (
	SelectorNamespace selectorType = iota
	SelectorPod
)

// TODO: make this private and expose a public conversion interface instead
type Converter struct{}

// VethNameForWorkload returns a deterministic veth name
// for the given Kubernetes workload (WEP) name and namespace.
func VethNameForWorkload(namespace, podname string) string {
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

// ParseWorkloadName extracts the Node name, Orchestrator, Pod name and endpoint from the
// given WorkloadEndpoint name.
// The expected format for k8s is <node>-k8s-<pod>-<endpoint>
func (c Converter) ParseWorkloadEndpointName(workloadName string) (names.WorkloadEndpointIdentifiers, error) {
	return names.ParseWorkloadEndpointName(workloadName)
}

// NamespaceToProfile converts a Namespace to a Calico Profile.  The Profile stores
// labels from the Namespace which are inherited by the WorkloadEndpoints within
// the Profile. This Profile also has the default ingress and egress rules, which are both 'allow'.
func (c Converter) NamespaceToProfile(ns *kapiv1.Namespace) (*model.KVPair, error) {
	// Generate the labels to apply to the profile, using a special prefix
	// to indicate that these are the labels from the parent Kubernetes Namespace.
	labels := map[string]string{}
	for k, v := range ns.Labels {
		labels[NamespaceLabelPrefix+k] = v
	}

	// Create the profile object.
	name := NamespaceProfileNamePrefix + ns.Name
	profile := apiv3.NewProfile()
	profile.ObjectMeta = metav1.ObjectMeta{
		Name:              name,
		CreationTimestamp: ns.CreationTimestamp,
		UID:               ns.UID,
	}
	profile.Spec = apiv3.ProfileSpec{
		Ingress: []apiv3.Rule{{Action: apiv3.Allow}},
		Egress:  []apiv3.Rule{{Action: apiv3.Allow}},
	}

	// Only set labels to apply when there are actually labels. This makes the
	// result of this function consistent with the struct as loaded directly
	// from etcd, which uses nil for the empty map.
	if len(labels) != 0 {
		profile.Spec.LabelsToApply = labels
	}

	// Embed the profile in a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name: name,
			Kind: apiv3.KindProfile,
		},
		Value:    profile,
		Revision: c.JoinProfileRevisions(ns.ResourceVersion, ""),
	}
	return &kvp, nil
}

// IsValidCalicoWorkloadEndpoint returns true if the pod should be shown as a workloadEndpoint
// in the Calico API and false otherwise.  Note: since we completely ignore notifications for
// invalid Pods, it is important that pods can only transition from not-valid to valid and not
// the other way.  If they transition from valid to invalid, we'll fail to emit a deletion
// event in the watcher.
func (c Converter) IsValidCalicoWorkloadEndpoint(pod *kapiv1.Pod) bool {
	if c.IsHostNetworked(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is host networked.")
		return false
	} else if !c.IsScheduled(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not scheduled.")
		return false
	}
	return true
}

// IsReadyCalicoPod returns true if the pod is a valid Calico WorkloadEndpoint and has
// an IP address assigned (i.e. it's ready for Calico networking).
func (c Converter) IsReadyCalicoPod(pod *kapiv1.Pod) bool {
	if !c.IsValidCalicoWorkloadEndpoint(pod) {
		return false
	} else if !c.HasIPAddress(pod) {
		log.WithField("pod", pod.Name).Debug("Pod does not have an IP address.")
		return false
	}
	return true
}

const (
	// Completed is documented but doesn't seem to be in the API, it should be safe to include.
	// Maybe it's in an older version of the API?
	podCompleted kapiv1.PodPhase = "Completed"
)

func (c Converter) IsFinished(pod *kapiv1.Pod) bool {
	switch pod.Status.Phase {
	case kapiv1.PodFailed, kapiv1.PodSucceeded, podCompleted:
		return true
	}
	return false
}

func (c Converter) IsScheduled(pod *kapiv1.Pod) bool {
	return pod.Spec.NodeName != ""
}

func (c Converter) IsHostNetworked(pod *kapiv1.Pod) bool {
	return pod.Spec.HostNetwork
}

func (c Converter) HasIPAddress(pod *kapiv1.Pod) bool {
	return pod.Status.PodIP != "" || pod.Annotations[AnnotationPodIP] != ""
}

// GetPodIPs extracts the IP addresses from a Kubernetes Pod.  At present, only a single IP
// is expected/supported.  GetPodIPs loads the IP either from the PodIP field, if present, or
// the calico podIP annotation.
func (c Converter) GetPodIPs(pod *kapiv1.Pod) ([]string, error) {
	var podIP string
	if podIP = pod.Status.PodIP; podIP != "" {
		log.WithField("ip", podIP).Debug("PodIP field filled in.")
	} else if podIP = pod.Annotations[AnnotationPodIP]; podIP != "" {
		log.WithField("ip", podIP).Debug("PodIP missing, falling back on Calico annotation.")
	} else {
		log.WithField("ip", podIP).Debug("Pod has no IP.")
		return nil, nil
	}
	_, ipNet, err := cnet.ParseCIDROrIP(podIP)
	if err != nil {
		log.WithFields(log.Fields{"ip": podIP, "pod": pod.Name}).WithError(err).Error("Failed to parse pod IP")
		return nil, err
	}
	return []string{ipNet.String()}, nil
}

// PodToWorkloadEndpoint converts a Pod to a WorkloadEndpoint.  It assumes the calling code
// has verified that the provided Pod is valid to convert to a WorkloadEndpoint.
// PodToWorkloadEndpoint requires a Pods Name and Node Name to be populated. It will
// fail to convert from a Pod to WorkloadEndpoint otherwise.
func (c Converter) PodToWorkloadEndpoint(pod *kapiv1.Pod) (*model.KVPair, error) {
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

	ipNets, err := c.GetPodIPs(pod)
	if err != nil {
		// IP address was present but malformed in some way, handle as an explicit failure.
		return nil, err
	}

	if c.IsFinished(pod) {
		// Pod is finished but not yet deleted.  In this state the IP will have been freed and returned to the pool
		// so we need to make sure we don't let the caller believe it still belongs to this endpoint.
		// Pods with no IPs will get filtered out before they get to Felix in the watcher syncer cache layer.
		// We can't pretend the workload endpoint is deleted _here_ because that would confuse users of the
		// native v3 Watch() API.
		ipNets = nil
	}

	// Generate the interface name based on workload.  This must match
	// the host-side veth configured by the CNI plugin.
	interfaceName := VethNameForWorkload(pod.Namespace, pod.Name)

	// Build the labels map.  Start with the pod labels, and append two additional labels for
	// namespace and orchestrator matches.
	labels := pod.Labels
	if labels == nil {
		labels = make(map[string]string, 2)
	}
	labels[apiv3.LabelNamespace] = pod.Namespace
	labels[apiv3.LabelOrchestrator] = apiv3.OrchestratorKubernetes

	if pod.Spec.ServiceAccountName != "" {
		labels[apiv3.LabelServiceAccount] = pod.Spec.ServiceAccountName
	}

	// Pull out floating IP annotation
	var floatingIPs []apiv3.IPNAT
	if annotation, ok := pod.Annotations["cni.projectcalico.org/floatingIPs"]; ok && len(ipNets) > 0 {

		// Parse Annotation data
		var ips []string
		err := json.Unmarshal([]byte(annotation), &ips)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s' as JSON: %s", annotation, err)
		}

		// Get target for NAT
		podip, podnet, err := cnet.ParseCIDROrIP(ipNets[0])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse pod IP: %s", err)
		}

		netmask, _ := podnet.Mask.Size()

		if netmask != 32 && netmask != 128 {
			return nil, fmt.Errorf("PodIP is not a valid IP: Mask size is %d, not 32 or 128", netmask)
		}

		for _, ip := range ips {
			floatingIPs = append(floatingIPs, apiv3.IPNAT{
				InternalIP: podip.String(),
				ExternalIP: ip,
			})
		}
	}

	// Map any named ports through.
	var endpointPorts []apiv3.EndpointPort
	for _, container := range pod.Spec.Containers {
		for _, containerPort := range container.Ports {
			if containerPort.Name != "" && containerPort.ContainerPort != 0 {
				var modelProto numorstring.Protocol
				switch containerPort.Protocol {
				case kapiv1.ProtocolUDP:
					modelProto = numorstring.ProtocolFromString("udp")
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

				endpointPorts = append(endpointPorts, apiv3.EndpointPort{
					Name:     containerPort.Name,
					Protocol: modelProto,
					Port:     uint16(containerPort.ContainerPort),
				})
			}
		}
	}

	// Create the workload endpoint.
	wep := apiv3.NewWorkloadEndpoint()
	wep.ObjectMeta = metav1.ObjectMeta{
		Name:              wepName,
		Namespace:         pod.Namespace,
		CreationTimestamp: pod.CreationTimestamp,
		UID:               pod.UID,
		Labels:            labels,
		GenerateName:      pod.GenerateName,
	}
	wep.Spec = apiv3.WorkloadEndpointSpec{
		Orchestrator:  "k8s",
		Node:          pod.Spec.NodeName,
		Pod:           pod.Name,
		Endpoint:      "eth0",
		InterfaceName: interfaceName,
		Profiles:      profiles,
		IPNetworks:    ipNets,
		Ports:         endpointPorts,
		IPNATs:        floatingIPs,
	}

	// Embed the workload endpoint into a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name:      wepName,
			Namespace: pod.Namespace,
			Kind:      apiv3.KindWorkloadEndpoint,
		},
		Value:    wep,
		Revision: pod.ResourceVersion,
	}
	return &kvp, nil
}

// K8sNetworkPolicyToCalico converts a k8s NetworkPolicy to a model.KVPair.
func (c Converter) K8sNetworkPolicyToCalico(np *networkingv1.NetworkPolicy) (*model.KVPair, error) {
	// Pull out important fields.
	policyName := fmt.Sprintf(K8sNetworkPolicyNamePrefix + np.Name)

	// We insert all the NetworkPolicy Policies at order 1000.0 after conversion.
	// This order might change in future.
	order := float64(1000.0)

	// Generate the ingress rules list.
	var ingressRules []apiv3.Rule
	for _, r := range np.Spec.Ingress {
		rules, err := c.k8sRuleToCalico(r.From, r.Ports, np.Namespace, true)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
		} else {
			ingressRules = append(ingressRules, rules...)
		}
	}

	// Generate the egress rules list.
	var egressRules []apiv3.Rule
	for _, r := range np.Spec.Egress {
		rules, err := c.k8sRuleToCalico(r.To, r.Ports, np.Namespace, false)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted")
		} else {
			egressRules = append(egressRules, rules...)
		}
	}

	// Calculate Types setting.
	ingress := false
	egress := false
	for _, policyType := range np.Spec.PolicyTypes {
		switch policyType {
		case networkingv1.PolicyTypeIngress:
			ingress = true
		case networkingv1.PolicyTypeEgress:
			egress = true
		}
	}
	types := []apiv3.PolicyType{}
	if ingress {
		types = append(types, apiv3.PolicyTypeIngress)
	}
	if egress {
		types = append(types, apiv3.PolicyTypeEgress)
	} else if len(egressRules) > 0 {
		// Egress was introduced at the same time as policyTypes.  It shouldn't be possible to
		// receive a NetworkPolicy with an egress rule but without "Egress" specified in its types,
		// but we'll warn about it anyway.
		log.Warn("K8s PolicyTypes don't include 'egress', but NetworkPolicy has egress rules.")
	}

	// If no types were specified in the policy, then we're running on a cluster that doesn't
	// include support for that field in the API.  In that case, the correct behavior is for the policy
	// to apply to only ingress traffic.
	if len(types) == 0 {
		types = append(types, apiv3.PolicyTypeIngress)
	}

	// Create the NetworkPolicy.
	policy := apiv3.NewNetworkPolicy()
	policy.ObjectMeta = metav1.ObjectMeta{
		Name:              policyName,
		Namespace:         np.Namespace,
		CreationTimestamp: np.CreationTimestamp,
		UID:               np.UID,
	}
	policy.Spec = apiv3.NetworkPolicySpec{
		Order:    &order,
		Selector: c.k8sSelectorToCalico(&np.Spec.PodSelector, SelectorPod),
		Ingress:  ingressRules,
		Egress:   egressRules,
		Types:    types,
	}

	// Build and return the KVPair.
	return &model.KVPair{
		Key: model.ResourceKey{
			Name:      policyName,
			Namespace: np.Namespace,
			Kind:      apiv3.KindNetworkPolicy,
		},
		Value:    policy,
		Revision: np.ResourceVersion,
	}, nil
}

// k8sSelectorToCalico takes a namespaced k8s label selector and returns the Calico
// equivalent.
func (c Converter) k8sSelectorToCalico(s *metav1.LabelSelector, selectorType selectorType) string {
	// Only prefix pod selectors - this won't work for namespace selectors.
	selectors := []string{}
	if selectorType == SelectorPod {
		selectors = append(selectors, fmt.Sprintf("%s == 'k8s'", apiv3.LabelOrchestrator))
	}

	if s == nil {
		return strings.Join(selectors, " && ")
	}

	// For namespace selectors, if they are present but have no terms, it means "select all
	// namespaces". We use empty string to represent the nil namespace selector, so use all() to
	// represent all namespaces.
	if selectorType == SelectorNamespace && len(s.MatchLabels) == 0 && len(s.MatchExpressions) == 0 {
		return "all()"
	}

	// matchLabels is a map key => value, it means match if (label[key] ==
	// value) for all keys.
	keys := make([]string, 0, len(s.MatchLabels))
	for k := range s.MatchLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.MatchLabels[k]
		selectors = append(selectors, fmt.Sprintf("%s == '%s'", k, v))
	}

	// matchExpressions is a list of in/notin/exists/doesnotexist tests.
	for _, e := range s.MatchExpressions {
		valueList := strings.Join(e.Values, "', '")

		// Each selector is formatted differently based on the operator.
		switch e.Operator {
		case metav1.LabelSelectorOpIn:
			selectors = append(selectors, fmt.Sprintf("%s in { '%s' }", e.Key, valueList))
		case metav1.LabelSelectorOpNotIn:
			selectors = append(selectors, fmt.Sprintf("%s not in { '%s' }", e.Key, valueList))
		case metav1.LabelSelectorOpExists:
			selectors = append(selectors, fmt.Sprintf("has(%s)", e.Key))
		case metav1.LabelSelectorOpDoesNotExist:
			selectors = append(selectors, fmt.Sprintf("! has(%s)", e.Key))
		}
	}

	return strings.Join(selectors, " && ")
}

func (c Converter) k8sRuleToCalico(rPeers []networkingv1.NetworkPolicyPeer, rPorts []networkingv1.NetworkPolicyPort, ns string, ingress bool) ([]apiv3.Rule, error) {
	rules := []apiv3.Rule{}
	peers := []*networkingv1.NetworkPolicyPeer{}
	ports := []*networkingv1.NetworkPolicyPort{}

	// Built up a list of the sources and a list of the destinations.
	for _, f := range rPeers {
		// We need to add a copy of the peer so all the rules don't
		// point to the same location.
		peers = append(peers, &networkingv1.NetworkPolicyPeer{
			NamespaceSelector: f.NamespaceSelector,
			PodSelector:       f.PodSelector,
			IPBlock:           f.IPBlock,
		})
	}
	for _, p := range rPorts {
		// We need to add a copy of the port so all the rules don't
		// point to the same location.
		port := networkingv1.NetworkPolicyPort{}
		if p.Port != nil {
			portval := intstr.FromString(p.Port.String())
			port.Port = &portval

			// TCP is the implicit default (as per the definition of NetworkPolicyPort).
			// Make the default explicit here because our data-model always requires
			// the protocol to be specified if we're doing a port match.
			port.Protocol = &protoTCP
		}
		if p.Protocol != nil {
			protval := kapiv1.Protocol(fmt.Sprintf("%s", *p.Protocol))
			port.Protocol = &protval
		}
		ports = append(ports, &port)
	}

	// If there no peers, or no ports, represent that as nil.
	if len(peers) == 0 {
		peers = []*networkingv1.NetworkPolicyPeer{nil}
	}
	if len(ports) == 0 {
		ports = []*networkingv1.NetworkPolicyPort{nil}
	}

	// Combine destinations with sources to generate rules.
	// TODO: This currently creates a lot of rules by making every combination of from / ports
	// into a rule.  We can combine these so that we don't need as many rules!
	for _, port := range ports {
		protocol, calicoPorts, err := c.k8sPortToCalicoFields(port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse k8s port: %s", err)
		}

		for _, peer := range peers {
			selector, nsSelector, nets, notNets := c.k8sPeerToCalicoFields(peer, ns)
			if ingress {
				// Build inbound rule and append to list.
				rules = append(rules, apiv3.Rule{
					Action:   "Allow",
					Protocol: protocol,
					Source: apiv3.EntityRule{
						Selector:          selector,
						NamespaceSelector: nsSelector,
						Nets:              nets,
						NotNets:           notNets,
					},
					Destination: apiv3.EntityRule{
						Ports: calicoPorts,
					},
				})
			} else {
				// Build outbound rule and append to list.
				rules = append(rules, apiv3.Rule{
					Action:   "Allow",
					Protocol: protocol,
					Destination: apiv3.EntityRule{
						Ports:             calicoPorts,
						Selector:          selector,
						NamespaceSelector: nsSelector,
						Nets:              nets,
						NotNets:           notNets,
					},
				})
			}
		}
	}
	return rules, nil
}

func (c Converter) k8sPortToCalicoFields(port *networkingv1.NetworkPolicyPort) (protocol *numorstring.Protocol, dstPorts []numorstring.Port, err error) {
	// If no port info, return zero values for all fields (protocol, dstPorts).
	if port == nil {
		return
	}
	// Port information available.
	dstPorts, err = c.k8sPortToCalico(*port)
	if err != nil {
		return
	}
	protocol = c.k8sProtocolToCalico(port.Protocol)
	return
}

func (c Converter) k8sProtocolToCalico(protocol *kapiv1.Protocol) *numorstring.Protocol {
	if protocol != nil {
		p := numorstring.ProtocolFromString(string(*protocol))
		return &p
	}
	return nil
}

func (c Converter) k8sPeerToCalicoFields(peer *networkingv1.NetworkPolicyPeer, ns string) (selector, nsSelector string, nets []string, notNets []string) {
	// If no peer, return zero values for all fields (selector, nets and !nets).
	if peer == nil {
		return
	}
	// Peer information available.
	// Determine the source selector for the rule.
	if peer.IPBlock != nil {
		// Convert the CIDR to include.
		_, ipNet, err := cnet.ParseCIDR(peer.IPBlock.CIDR)
		if err != nil {
			log.WithField("cidr", peer.IPBlock.CIDR).WithError(err).Error("Failed to parse CIDR")
			return
		}
		nets = []string{ipNet.String()}

		// Convert the CIDRs to exclude.
		for _, exception := range peer.IPBlock.Except {
			_, ipNet, err = cnet.ParseCIDR(exception)
			if err != nil {
				log.WithField("cidr", exception).WithError(err).Error("Failed to parse CIDR")
				return
			}
			notNets = append(notNets, ipNet.String())
		}
		// If IPBlock is set, then PodSelector and NamespaceSelector cannot be.
		return
	}

	// IPBlock is not set to get here.
	// Note that k8sSelectorToCalico() accepts nil values of the selector.
	selector = c.k8sSelectorToCalico(peer.PodSelector, SelectorPod)
	nsSelector = c.k8sSelectorToCalico(peer.NamespaceSelector, SelectorNamespace)
	return
}

func (c Converter) k8sPortToCalico(port networkingv1.NetworkPolicyPort) ([]numorstring.Port, error) {
	var portList []numorstring.Port
	if port.Port != nil {
		p, err := numorstring.PortFromString(port.Port.String())
		if err != nil {
			return nil, fmt.Errorf("invalid port %+v: %s", port.Port, err)
		}
		return append(portList, p), nil
	}

	// No ports - return empty list.
	return portList, nil
}

// ProfileNameToNamespace extracts the Namespace name from the given Profile name.
func (c Converter) ProfileNameToNamespace(profileName string) (string, error) {
	// Profile objects backed by Namespaces have form "kns.<ns_name>"
	if !strings.HasPrefix(profileName, NamespaceProfileNamePrefix) {
		// This is not backed by a Kubernetes Namespace.
		return "", fmt.Errorf("Profile %s not backed by a Namespace", profileName)
	}

	return strings.TrimPrefix(profileName, NamespaceProfileNamePrefix), nil
}

// JoinNetworkPolicyRevisions constructs the revision from the individual CRD and K8s NetworkPolicy
// revisions.
func (c Converter) JoinNetworkPolicyRevisions(crdNPRev, k8sNPRev string) string {
	return crdNPRev + "/" + k8sNPRev
}

// SplitNetworkPolicyRevision extracts the CRD and K8s NetworkPolicy revisions from the combined
// revision returned on the KDD NetworkPolicy client.
func (c Converter) SplitNetworkPolicyRevision(rev string) (crdNPRev string, k8sNPRev string, err error) {
	if rev == "" {
		return
	}

	revs := strings.Split(rev, "/")
	if len(revs) != 2 {
		err = fmt.Errorf("ResourceVersion is not valid: %s", rev)
		return
	}

	crdNPRev = revs[0]
	k8sNPRev = revs[1]
	return
}

// serviceAccountNameToProfileName creates a profile name that is a join
// of 'ksa.' + namespace + "." + serviceaccount name.
func serviceAccountNameToProfileName(sa, namespace string) string {
	// Need to incorporate the namespace into the name of the sa based profile
	// to make them globally unique
	if namespace == "" {
		namespace = "default"
	}
	return ServiceAccountProfileNamePrefix + namespace + "." + sa
}

// ServiceAccountToProfile converts a ServiceAccount to a Calico Profile.  The Profile stores
// labels from the ServiceAccount which are inherited by the WorkloadEndpoints within
// the Profile.
func (c Converter) ServiceAccountToProfile(sa *kapiv1.ServiceAccount) (*model.KVPair, error) {
	// Generate the labels to apply to the profile, using a special prefix
	// to indicate that these are the labels from the parent Kubernetes ServiceAccount.
	labels := map[string]string{}
	for k, v := range sa.ObjectMeta.Labels {
		labels[ServiceAccountLabelPrefix+k] = v
	}

	name := serviceAccountNameToProfileName(sa.Name, sa.Namespace)
	profile := apiv3.NewProfile()
	profile.ObjectMeta = metav1.ObjectMeta{
		Name:              name,
		CreationTimestamp: sa.CreationTimestamp,
		UID:               sa.UID,
	}

	// Only set labels to apply when there are actually labels. This makes the
	// result of this function consistent with the struct as loaded directly
	// from etcd, which uses nil for the empty map.
	if len(labels) != 0 {
		profile.Spec.LabelsToApply = labels
	} else {
		profile.Spec.LabelsToApply = nil
	}

	// Embed the profile in a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name: name,
			Kind: apiv3.KindProfile,
		},
		Value:    profile,
		Revision: c.JoinProfileRevisions("", sa.ResourceVersion),
	}
	return &kvp, nil
}

// ProfileNameToServiceAccount extracts the ServiceAccount name from the given Profile name.
func (c Converter) ProfileNameToServiceAccount(profileName string) (ns, sa string, err error) {

	// Profile objects backed by ServiceAccounts have form "ksa.<namespace>.<sa_name>"
	if !strings.HasPrefix(profileName, ServiceAccountProfileNamePrefix) {
		// This is not backed by a Kubernetes ServiceAccount.
		err = fmt.Errorf("Profile %s not backed by a ServiceAccount", profileName)
		return
	}

	names := strings.SplitN(profileName, ".", 3)
	if len(names) != 3 {
		err = fmt.Errorf("Profile %s is not formatted correctly", profileName)
		return
	}

	ns = names[1]
	sa = names[2]
	return
}

// JoinProfileRevisions constructs the revision from the individual namespace and serviceaccount
// revisions.
// This is conditional on the feature flag for serviceaccount set or not.
func (c Converter) JoinProfileRevisions(nsRev, saRev string) string {
	return nsRev + "/" + saRev
}

// SplitProfileRevision extracts the namespace and serviceaccount revisions from the combined
// revision returned on the KDD service account based profile.
// This is conditional on the feature flag for serviceaccount set or not.
func (c Converter) SplitProfileRevision(rev string) (nsRev string, saRev string, err error) {
	if rev == "" {
		return
	}

	revs := strings.Split(rev, "/")
	if len(revs) != 2 {
		err = fmt.Errorf("ResourceVersion is not valid: %s", rev)
		return
	}

	nsRev = revs[0]
	saRev = revs[1]
	return
}
