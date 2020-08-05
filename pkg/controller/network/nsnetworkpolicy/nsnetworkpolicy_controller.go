/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nsnetworkpolicy

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	workspacev1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	ksnetclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/network/v1alpha1"
	nspolicy "kubesphere.io/kubesphere/pkg/client/informers/externalversions/network/v1alpha1"
	workspace "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/network"
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
	options "kubesphere.io/kubesphere/pkg/simple/client/network"
)

const (
	//TODO  use set to track service:map
	//use period sync service label in NSNP
	defaultSleepDuration = 60 * time.Second

	defaultThread = 5
	defaultSync   = "5m"

	//whether network isolate is enable in namespace
	NamespaceNPAnnotationKey     = "kubesphere.io/network-isolate"
	NamespaceNPAnnotationEnabled = "enabled"

	NodeNSNPAnnotationKey = "kubesphere.io/snat-node-ips"

	AnnotationNPNAME = network.NSNPPrefix + "network-isolate"

	//TODO: configure it
	DNSLocalIP        = "169.254.25.10"
	DNSPort           = 53
	DNSNamespace      = "kube-system"
	DNSServiceName    = "kube-dns"
	DNSServiceCoreDNS = "coredns"
)

// namespacenpController implements the Controller interface for managing kubesphere network policies
// and convery them to k8s NetworkPolicies, then syncing them to the provider.
type NSNetworkPolicyController struct {
	client         kubernetes.Interface
	ksclient       ksnetclient.NetworkV1alpha1Interface
	informer       nspolicy.NamespaceNetworkPolicyInformer
	informerSynced cache.InformerSynced

	//TODO support service in same namespace
	serviceInformer       v1.ServiceInformer
	serviceInformerSynced cache.InformerSynced

	nodeInformer       v1.NodeInformer
	nodeInformerSynced cache.InformerSynced

	workspaceInformer       workspace.WorkspaceInformer
	workspaceInformerSynced cache.InformerSynced

	namespaceInformer       v1.NamespaceInformer
	namespaceInformerSynced cache.InformerSynced

	provider provider.NsNetworkPolicyProvider
	options  options.NSNPOptions

	nsQueue   workqueue.RateLimitingInterface
	nsnpQueue workqueue.RateLimitingInterface
}

func stringToCIDR(ipStr string) (string, error) {
	cidr := ""
	if ip := net.ParseIP(ipStr); ip != nil {
		if ip.To4() != nil {
			cidr = ipStr + "/32"
		} else {
			cidr = ipStr + "/128"
		}
	} else {
		return cidr, fmt.Errorf("ip string  %s parse error\n", ipStr)
	}

	return cidr, nil
}

func generateDNSRule(nameServers []string) (netv1.NetworkPolicyEgressRule, error) {
	var rule netv1.NetworkPolicyEgressRule
	for _, nameServer := range nameServers {
		cidr, err := stringToCIDR(nameServer)
		if err != nil {
			return rule, err
		}
		rule.To = append(rule.To, netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: cidr,
			},
		})
	}

	protocolTCP := corev1.ProtocolTCP
	protocolUDP := corev1.ProtocolUDP
	dnsPort := intstr.FromInt(DNSPort)
	rule.Ports = append(rule.Ports, netv1.NetworkPolicyPort{
		Protocol: &protocolTCP,
		Port:     &dnsPort,
	}, netv1.NetworkPolicyPort{
		Protocol: &protocolUDP,
		Port:     &dnsPort,
	})

	return rule, nil
}

func (c *NSNetworkPolicyController) generateDNSServiceRule() (netv1.NetworkPolicyEgressRule, error) {
	peer, ports, err := c.handlerPeerService(DNSNamespace, DNSServiceName, false)
	if err != nil {
		peer, ports, err = c.handlerPeerService(DNSNamespace, DNSServiceCoreDNS, false)
	}

	return netv1.NetworkPolicyEgressRule{
		Ports: ports,
		To:    []netv1.NetworkPolicyPeer{peer},
	}, err
}

func (c *NSNetworkPolicyController) handlerPeerService(namespace string, name string, ingress bool) (netv1.NetworkPolicyPeer, []netv1.NetworkPolicyPort, error) {
	peerNP := netv1.NetworkPolicyPeer{}
	var ports []netv1.NetworkPolicyPort

	service, err := c.serviceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		return peerNP, nil, err
	}

	peerNP.PodSelector = new(metav1.LabelSelector)
	peerNP.NamespaceSelector = new(metav1.LabelSelector)

	if len(service.Spec.Selector) <= 0 {
		return peerNP, nil, fmt.Errorf("service %s/%s has no podselect", namespace, name)
	}

	peerNP.PodSelector.MatchLabels = make(map[string]string)
	for key, value := range service.Spec.Selector {
		peerNP.PodSelector.MatchLabels[key] = value
	}
	peerNP.NamespaceSelector.MatchLabels = make(map[string]string)
	peerNP.NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = namespace

	//only allow traffic to service exposed ports
	if !ingress {
		ports = make([]netv1.NetworkPolicyPort, 0)
		for _, port := range service.Spec.Ports {
			protocol := port.Protocol
			portIntString := intstr.FromInt(int(port.Port))
			ports = append(ports, netv1.NetworkPolicyPort{
				Protocol: &protocol,
				Port:     &portIntString,
			})
		}
	}

	return peerNP, ports, err
}

func (c *NSNetworkPolicyController) convertPeer(peer v1alpha1.NetworkPolicyPeer, ingress bool) (netv1.NetworkPolicyPeer, []netv1.NetworkPolicyPort, error) {
	peerNP := netv1.NetworkPolicyPeer{}
	var ports []netv1.NetworkPolicyPort

	if peer.ServiceSelector != nil {
		namespace := peer.ServiceSelector.Namespace
		name := peer.ServiceSelector.Name

		return c.handlerPeerService(namespace, name, ingress)
	} else if peer.NamespaceSelector != nil {
		name := peer.NamespaceSelector.Name

		peerNP.NamespaceSelector = new(metav1.LabelSelector)
		peerNP.NamespaceSelector.MatchLabels = make(map[string]string)
		peerNP.NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = name
	} else if peer.IPBlock != nil {
		peerNP.IPBlock = peer.IPBlock
	} else {
		klog.Errorf("Invalid nsnp peer %v\n", peer)
		return peerNP, nil, fmt.Errorf("Invalid nsnp peer %v\n", peer)
	}

	return peerNP, ports, nil
}

func (c *NSNetworkPolicyController) convertToK8sNP(n *v1alpha1.NamespaceNetworkPolicy) (*netv1.NetworkPolicy, error) {
	np := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      network.NSNPPrefix + n.Name,
			Namespace: n.Namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: make([]netv1.PolicyType, 0),
		},
	}

	if n.Spec.Egress != nil {
		np.Spec.Egress = make([]netv1.NetworkPolicyEgressRule, 0)
		for _, egress := range n.Spec.Egress {
			tmpRule := netv1.NetworkPolicyEgressRule{}
			for _, peer := range egress.To {
				peer, ports, err := c.convertPeer(peer, false)
				if err != nil {
					return nil, err
				}
				if ports != nil {
					np.Spec.Egress = append(np.Spec.Egress, netv1.NetworkPolicyEgressRule{
						Ports: ports,
						To:    []netv1.NetworkPolicyPeer{peer},
					})
					continue
				}
				tmpRule.To = append(tmpRule.To, peer)
			}
			tmpRule.Ports = egress.Ports
			if tmpRule.To == nil {
				continue
			}
			np.Spec.Egress = append(np.Spec.Egress, tmpRule)
		}
		np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, netv1.PolicyTypeEgress)
	}

	if n.Spec.Ingress != nil {
		np.Spec.Ingress = make([]netv1.NetworkPolicyIngressRule, 0)
		for _, ingress := range n.Spec.Ingress {
			tmpRule := netv1.NetworkPolicyIngressRule{}
			for _, peer := range ingress.From {
				peer, ports, err := c.convertPeer(peer, true)
				if err != nil {
					return nil, err
				}
				if ports != nil {
					np.Spec.Ingress = append(np.Spec.Ingress, netv1.NetworkPolicyIngressRule{
						Ports: ports,
						From:  []netv1.NetworkPolicyPeer{peer},
					})
				}
				tmpRule.From = append(tmpRule.From, peer)
			}
			tmpRule.Ports = ingress.Ports
			np.Spec.Ingress = append(np.Spec.Ingress, tmpRule)
		}
		np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, netv1.PolicyTypeIngress)
	}

	return np, nil
}

func (c *NSNetworkPolicyController) generateNodeRule() (netv1.NetworkPolicyIngressRule, error) {
	var (
		rule netv1.NetworkPolicyIngressRule
		ips  []string
	)

	nodes, err := c.nodeInformer.Lister().List(labels.Everything())
	if err != nil {
		return rule, err
	}
	for _, node := range nodes {
		snatIPs := node.Annotations[NodeNSNPAnnotationKey]
		if snatIPs != "" {
			ips = append(ips, strings.Split(snatIPs, ";")...)
		}
	}

	sort.Strings(ips)

	for _, ip := range ips {
		cidr, err := stringToCIDR(ip)
		if err != nil {
			continue
		}

		rule.From = append(rule.From, netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: cidr,
			},
		})
	}

	return rule, nil
}

func (c *NSNetworkPolicyController) generateNSNP(workspace string, namespace string, matchWorkspace bool) *netv1.NetworkPolicy {
	policy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AnnotationNPNAME,
			Namespace: namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: make([]netv1.PolicyType, 0),
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}},
			}},
		},
	}

	policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, netv1.PolicyTypeIngress)

	if matchWorkspace {
		policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[constants.WorkspaceLabelKey] = workspace
	} else {
		policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = namespace
	}

	for _, allowedIngressNamespace := range c.options.AllowedIngressNamespaces {
		defaultAllowedIngress := netv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.NamespaceLabelKey: allowedIngressNamespace,
				},
			},
		}
		policy.Spec.Ingress[0].From = append(policy.Spec.Ingress[0].From, defaultAllowedIngress)
	}

	return policy
}

func (c *NSNetworkPolicyController) nsEnqueue(ns *corev1.Namespace) {
	key, err := cache.MetaNamespaceKeyFunc(ns)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("Get namespace key %s failed", ns.Name))
		return
	}

	workspaceName := ns.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return
	}

	c.nsQueue.Add(key)
}

func (c *NSNetworkPolicyController) addWorkspace(newObj interface{}) {
	new := newObj.(*workspacev1alpha1.Workspace)

	label := labels.SelectorFromSet(labels.Set{constants.WorkspaceLabelKey: new.Name})
	nsList, err := c.namespaceInformer.Lister().List(label)
	if err != nil {
		klog.Errorf("Error while list namespace by label %s", label.String())
		return
	}

	for _, ns := range nsList {
		c.nsEnqueue(ns)
	}
}

func (c *NSNetworkPolicyController) addNode(newObj interface{}) {
	nsList, err := c.namespaceInformer.Lister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Error while list namespace by label")
		return
	}

	for _, ns := range nsList {
		c.nsEnqueue(ns)
	}
}

func (c *NSNetworkPolicyController) addNamespace(obj interface{}) {
	ns := obj.(*corev1.Namespace)

	workspaceName := ns.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return
	}

	c.nsEnqueue(ns)
}

func namespaceNetworkIsolateEnabled(ns *corev1.Namespace) bool {
	if ns.Annotations != nil && ns.Annotations[NamespaceNPAnnotationKey] == NamespaceNPAnnotationEnabled {
		return true
	}

	return false
}

func (c *NSNetworkPolicyController) syncNs(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("Not a valid controller key %s, %#v", key, err)
		return err
	}

	ns, err := c.namespaceInformer.Lister().Get(name)
	if err != nil {
		// not found, possibly been deleted
		if errors.IsNotFound(err) {
			klog.V(2).Infof("Namespace %v has been deleted", key)
			return nil
		}

		return err
	}

	workspaceName := ns.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}

	wksp, err := c.workspaceInformer.Lister().Get(workspaceName)
	if err != nil {
		//Should not be here
		if errors.IsNotFound(err) {
			klog.V(2).Infof("Workspace %v has been deleted", workspaceName)
			return nil
		}

		return err
	}

	matchWorkspace := false
	delete := false
	nsnpList, err := c.informer.Lister().NamespaceNetworkPolicies(ns.Name).List(labels.Everything())
	if namespaceNetworkIsolateEnabled(ns) {
		matchWorkspace = false
	} else if workspaceNetworkIsolationEnabled(wksp) {
		matchWorkspace = true
	} else {
		delete = true
	}
	if delete || matchWorkspace {
		//delete all namespace np when networkisolate not active
		if err == nil && len(nsnpList) > 0 {
			if c.ksclient.NamespaceNetworkPolicies(ns.Name).DeleteCollection(nil, typev1.ListOptions{}) != nil {
				klog.Errorf("Error when delete all nsnps in namespace %s", ns.Name)
			}
		}
	}

	policy := c.generateNSNP(workspaceName, ns.Name, matchWorkspace)
	if shouldAddDNSRule(nsnpList) {
		ruleDNS, err := generateDNSRule([]string{DNSLocalIP})
		if err != nil {
			return err
		}
		policy.Spec.Egress = append(policy.Spec.Egress, ruleDNS)
		ruleDNSService, err := c.generateDNSServiceRule()
		if err == nil {
			policy.Spec.Egress = append(policy.Spec.Egress, ruleDNSService)
		} else {
			klog.Warningf("Cannot handle service %s or %s", DNSServiceName, DNSServiceCoreDNS)
		}
		policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, netv1.PolicyTypeEgress)
	}
	ruleNode, err := c.generateNodeRule()
	if err != nil {
		return err
	}
	if len(ruleNode.From) > 0 {
		policy.Spec.Ingress = append(policy.Spec.Ingress, ruleNode)
	}

	if delete {
		c.provider.Delete(c.provider.GetKey(AnnotationNPNAME, ns.Name))
	} else {
		err = c.provider.Set(policy)
		if err != nil {
			klog.Errorf("Error while converting %#v to provider's network policy.", policy)
			return err
		}
	}

	return nil
}

func shouldAddDNSRule(nsnpList []*v1alpha1.NamespaceNetworkPolicy) bool {
	for _, nsnp := range nsnpList {
		if len(nsnp.Spec.Egress) > 0 {
			return true
		}
	}

	return false
}

func (c *NSNetworkPolicyController) nsWorker() {
	for c.processNsWorkItem() {
	}
}

func (c *NSNetworkPolicyController) processNsWorkItem() bool {
	key, quit := c.nsQueue.Get()
	if quit {
		return false
	}
	defer c.nsQueue.Done(key)

	if err := c.syncNs(key.(string)); err != nil {
		klog.Errorf("Error when syncns %s", err)
	}

	return true
}

func (c *NSNetworkPolicyController) nsnpEnqueue(obj interface{}) {
	nsnp := obj.(*v1alpha1.NamespaceNetworkPolicy)

	key, err := cache.MetaNamespaceKeyFunc(nsnp)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("get namespace network policy key %s failed", nsnp.Name))
		return
	}

	c.nsnpQueue.Add(key)
}

func (c *NSNetworkPolicyController) syncNSNP(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("Not a valid controller key %s, %#v", key, err)
		return err
	}

	c.nsQueue.Add(namespace)

	nsnp, err := c.informer.Lister().NamespaceNetworkPolicies(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("NSNP %v has been deleted", key)
			c.provider.Delete(c.provider.GetKey(network.NSNPPrefix+name, namespace))
			return nil
		}

		return err
	}

	np, err := c.convertToK8sNP(nsnp)
	if err != nil {
		klog.Errorf("Error while convert nsnp to k8snp: %s", err)
		return err
	}
	err = c.provider.Set(np)
	if err != nil {
		klog.Errorf("Error while set provider: %s", err)
		return err
	}

	return nil
}

func (c *NSNetworkPolicyController) nsNPWorker() {
	for c.processNSNPWorkItem() {
	}
}

func (c *NSNetworkPolicyController) processNSNPWorkItem() bool {
	key, quit := c.nsnpQueue.Get()
	if quit {
		return false
	}
	defer c.nsnpQueue.Done(key)

	c.syncNSNP(key.(string))

	return true
}

func workspaceNetworkIsolationEnabled(wksp *workspacev1alpha1.Workspace) bool {
	if wksp.Spec.NetworkIsolation != nil && *wksp.Spec.NetworkIsolation {
		return true
	}
	return false
}

// NewnamespacenpController returns a controller which manages NSNSP objects.
func NewNSNetworkPolicyController(
	client kubernetes.Interface,
	ksclient ksnetclient.NetworkV1alpha1Interface,
	nsnpInformer nspolicy.NamespaceNetworkPolicyInformer,
	serviceInformer v1.ServiceInformer,
	nodeInformer v1.NodeInformer,
	workspaceInformer workspace.WorkspaceInformer,
	namespaceInformer v1.NamespaceInformer,
	policyProvider provider.NsNetworkPolicyProvider,
	options options.NSNPOptions) *NSNetworkPolicyController {

	controller := &NSNetworkPolicyController{
		client:                  client,
		ksclient:                ksclient,
		informer:                nsnpInformer,
		informerSynced:          nsnpInformer.Informer().HasSynced,
		serviceInformer:         serviceInformer,
		serviceInformerSynced:   serviceInformer.Informer().HasSynced,
		nodeInformer:            nodeInformer,
		nodeInformerSynced:      nodeInformer.Informer().HasSynced,
		workspaceInformer:       workspaceInformer,
		workspaceInformerSynced: workspaceInformer.Informer().HasSynced,
		namespaceInformer:       namespaceInformer,
		namespaceInformerSynced: namespaceInformer.Informer().HasSynced,
		provider:                policyProvider,
		nsQueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespace"),
		nsnpQueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespacenp"),
		options:                 options,
	}

	workspaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addWorkspace,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*workspacev1alpha1.Workspace)
			new := newObj.(*workspacev1alpha1.Workspace)
			if workspaceNetworkIsolationEnabled(old) == workspaceNetworkIsolationEnabled(new) {
				return
			}
			controller.addWorkspace(newObj)
		},
	})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addNode,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*corev1.Node)
			new := newObj.(*corev1.Node)
			if old.Annotations[NodeNSNPAnnotationKey] == new.Annotations[NodeNSNPAnnotationKey] {
				return
			}
			controller.addNode(newObj)
		},
	})

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addNamespace,
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			old := oldObj.(*corev1.Namespace)
			new := newObj.(*corev1.Namespace)
			if old.Annotations[NamespaceNPAnnotationKey] == new.Annotations[NamespaceNPAnnotationKey] {
				return
			}
			controller.addNamespace(newObj)
		},
	})

	nsnpInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.nsnpEnqueue(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			controller.nsnpEnqueue(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			controller.nsnpEnqueue(obj)
		},
	}, defaultSleepDuration)

	return controller
}

func (c *NSNetworkPolicyController) Start(stopCh <-chan struct{}) error {
	return c.Run(defaultThread, defaultSync, stopCh)
}

// Run starts the controller.
func (c *NSNetworkPolicyController) Run(threadiness int, reconcilerPeriod string, stopCh <-chan struct{}) error {
	defer uruntime.HandleCrash()

	defer c.nsQueue.ShutDown()
	defer c.nsnpQueue.ShutDown()

	klog.Info("Waiting to sync with Kubernetes API (NSNP)")
	if ok := cache.WaitForCacheSync(stopCh, c.informerSynced, c.serviceInformerSynced, c.workspaceInformerSynced, c.namespaceInformerSynced, c.nodeInformerSynced); !ok {
		return fmt.Errorf("Failed to wait for caches to sync")
	}
	klog.Info("Finished syncing with Kubernetes API (NSNP)")

	// Start a number of worker threads to read from the queue. Each worker
	// will pull keys off the resource cache event queue and sync them to the
	// K8s datastore.
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.nsWorker, time.Second, stopCh)
		go wait.Until(c.nsNPWorker, time.Second, stopCh)
	}

	//Work to sync K8s NetworkPolicy
	go c.provider.Start(stopCh)

	klog.Info("NSNP controller is now running")
	<-stopCh
	klog.Info("Stopping NSNP controller")

	return nil
}
