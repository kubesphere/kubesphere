package nsnetworkpolicy

import (
	"fmt"
	"net"
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
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
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

	AnnotationNPNAME = "network-isolate"
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

	workspaceInformer       workspace.WorkspaceInformer
	workspaceInformerSynced cache.InformerSynced

	namespaceInformer       v1.NamespaceInformer
	namespaceInformerSynced cache.InformerSynced

	provider provider.NsNetworkPolicyProvider

	nsQueue   workqueue.RateLimitingInterface
	nsnpQueue workqueue.RateLimitingInterface
}

func (c *NSNetworkPolicyController) convertPeer(peer v1alpha1.NetworkPolicyPeer, ingress bool) (netv1.NetworkPolicyPeer, []netv1.NetworkPolicyPort, error) {
	rule := netv1.NetworkPolicyPeer{}
	var ports []netv1.NetworkPolicyPort

	if peer.ServiceSelector != nil {
		namespace := peer.ServiceSelector.Namespace
		name := peer.ServiceSelector.Name
		service, err := c.serviceInformer.Lister().Services(namespace).Get(name)
		if err != nil {
			return rule, nil, err
		}
		if ingress {
			rule.PodSelector = new(metav1.LabelSelector)
			rule.NamespaceSelector = new(metav1.LabelSelector)

			if len(service.Spec.Selector) <= 0 {
				return rule, nil, fmt.Errorf("service %s/%s has no podselect", namespace, name)
			}

			rule.PodSelector.MatchLabels = make(map[string]string)
			for key, value := range service.Spec.Selector {
				rule.PodSelector.MatchLabels[key] = value
			}
			rule.NamespaceSelector.MatchLabels = make(map[string]string)
			rule.NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = namespace
		} else {
			//only allow to service clusterip and service ports
			cidr := ""
			if ip := net.ParseIP(service.Spec.ClusterIP); ip != nil {
				if ip.To4() != nil {
					cidr = service.Spec.ClusterIP + "/32"
				} else {
					cidr = service.Spec.ClusterIP + "/128"
				}
			} else {
				return rule, nil, fmt.Errorf("Service %s/%s ClusterIP  %s parse error\n", service.Namespace, service.Name, service.Spec.ClusterIP)
			}
			rule.IPBlock = &netv1.IPBlock{
				CIDR: cidr,
			}

			ports = make([]netv1.NetworkPolicyPort, 0)
			for _, port := range service.Spec.Ports {
				portIntString := intstr.FromInt(int(port.Port))
				ports = append(ports, netv1.NetworkPolicyPort{
					Protocol: &port.Protocol,
					Port:     &portIntString,
				})
			}
		}
	} else if peer.NamespaceSelector != nil {
		name := peer.NamespaceSelector.Name

		rule.NamespaceSelector = new(metav1.LabelSelector)
		rule.NamespaceSelector.MatchLabels = make(map[string]string)
		rule.NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = name
	} else if peer.IPBlock != nil {
		rule.IPBlock = peer.IPBlock
	} else {
		klog.Errorf("Invalid nsnp peer %v\n", peer)
		return rule, nil, fmt.Errorf("Invalid nsnp peer %v\n", peer)
	}

	return rule, ports, nil
}

func (c *NSNetworkPolicyController) convertToK8sNP(n *v1alpha1.NamespaceNetworkPolicy) (*netv1.NetworkPolicy, error) {
	np := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: make([]netv1.PolicyType, 0),
		},
	}

	if n.Spec.Egress != nil {
		np.Spec.Egress = make([]netv1.NetworkPolicyEgressRule, len(n.Spec.Egress))
		for indexEgress, egress := range n.Spec.Egress {
			for _, peer := range egress.To {
				rule, ports, err := c.convertPeer(peer, false)
				if err != nil {
					return nil, err
				}
				np.Spec.Egress[indexEgress].To = append(np.Spec.Egress[indexEgress].To, rule)
				np.Spec.Egress[indexEgress].Ports = append(np.Spec.Egress[indexEgress].Ports, ports...)
			}
			np.Spec.Egress[indexEgress].Ports = append(np.Spec.Egress[indexEgress].Ports, egress.Ports...)
		}
		np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, netv1.PolicyTypeEgress)
	}

	if n.Spec.Ingress != nil {
		np.Spec.Ingress = make([]netv1.NetworkPolicyIngressRule, len(n.Spec.Ingress))
		for indexIngress, ingress := range n.Spec.Ingress {
			for _, peer := range ingress.From {
				rule, ports, err := c.convertPeer(peer, true)
				if err != nil {
					return nil, err
				}
				np.Spec.Ingress[indexIngress].From = append(np.Spec.Ingress[indexIngress].From, rule)
				np.Spec.Ingress[indexIngress].Ports = append(np.Spec.Ingress[indexIngress].Ports, ports...)
			}
			np.Spec.Ingress[indexIngress].Ports = append(np.Spec.Ingress[indexIngress].Ports, ingress.Ports...)
		}
		np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, netv1.PolicyTypeIngress)
	}

	return np, nil
}

func generateNSNP(workspace string, namespace string, matchWorkspace bool) *netv1.NetworkPolicy {
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
			Egress: []netv1.NetworkPolicyEgressRule{{
				To: []netv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}},
			}},
		},
	}

	policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, netv1.PolicyTypeIngress, netv1.PolicyTypeEgress)

	if matchWorkspace {
		policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[constants.WorkspaceLabelKey] = workspace
		policy.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels[constants.WorkspaceLabelKey] = workspace
	} else {
		policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = namespace
		policy.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels[constants.NamespaceLabelKey] = namespace
	}

	return policy
}

func (c *NSNetworkPolicyController) nsEnqueue(ns *corev1.Namespace) {
	key, err := cache.MetaNamespaceKeyFunc(ns)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("Get namespace key %s failed", ns.Name))
		return
	}

	klog.V(4).Infof("Enqueue namespace %s", ns.Name)
	c.nsQueue.Add(key)
}

func (c *NSNetworkPolicyController) addWorkspace(newObj interface{}) {
	new := newObj.(*workspacev1alpha1.Workspace)

	klog.V(4).Infof("Add workspace %s", new.Name)

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

func (c *NSNetworkPolicyController) addNamespace(obj interface{}) {
	ns := obj.(*corev1.Namespace)

	workspaceName := ns.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return
	}

	klog.V(4).Infof("Add namespace %s", ns.Name)

	c.nsEnqueue(ns)
}

func isNetworkIsolateEnabled(ns *corev1.Namespace) bool {
	if ns.Annotations[NamespaceNPAnnotationKey] == NamespaceNPAnnotationEnabled {
		return true
	}

	return false
}

func hadNamespaceLabel(ns *corev1.Namespace) bool {
	if ns.Annotations[constants.NamespaceLabelKey] == ns.Name {
		return true
	}

	return false
}

func (c *NSNetworkPolicyController) syncNs(key string) error {
	klog.V(4).Infof("Sync namespace %s", key)

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
		klog.Error("Workspace name should not be empty")
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

	//Maybe some ns not labeled
	if !hadNamespaceLabel(ns) {
		ns.Labels[constants.NamespaceLabelKey] = ns.Name
		_, err := c.client.CoreV1().Namespaces().Update(ns)
		if err != nil {
			//Just log, label can also be added by namespace controller
			klog.Errorf("cannot label namespace %s", ns.Name)
		}
	}

	matchWorkspace := false
	delete := false
	if isNetworkIsolateEnabled(ns) {
		matchWorkspace = false
	} else if wksp.Spec.NetworkIsolation {
		matchWorkspace = true
	} else {
		delete = true
	}

	policy := generateNSNP(workspaceName, ns.Name, matchWorkspace)
	if delete {
		c.provider.Delete(c.provider.GetKey(AnnotationNPNAME, ns.Name))
		//delete all namespace np when networkisolate not active
		if c.ksclient.NamespaceNetworkPolicies(ns.Name).DeleteCollection(nil, typev1.ListOptions{}) != nil {
			klog.Errorf("Error when delete all nsnps in namespace %s", ns.Name)
		}
	} else {
		err = c.provider.Set(policy)
		if err != nil {
			klog.Errorf("Error while converting %#v to provider's network policy.", policy)
			return err
		}
	}

	return nil
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

	ns, err := c.namespaceInformer.Lister().Get(namespace)
	if !isNetworkIsolateEnabled(ns) {
		klog.Infof("Delete NSNP %s when namespace isolate is inactive", key)
		c.provider.Delete(c.provider.GetKey(name, namespace))
		return nil
	}

	nsnp, err := c.informer.Lister().NamespaceNetworkPolicies(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("NSNP %v has been deleted", key)
			c.provider.Delete(c.provider.GetKey(name, namespace))
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

// NewnamespacenpController returns a controller which manages NSNSP objects.
func NewNSNetworkPolicyController(
	client kubernetes.Interface,
	ksclient ksnetclient.NetworkV1alpha1Interface,
	nsnpInformer nspolicy.NamespaceNetworkPolicyInformer,
	serviceInformer v1.ServiceInformer,
	workspaceInformer workspace.WorkspaceInformer,
	namespaceInformer v1.NamespaceInformer,
	policyProvider provider.NsNetworkPolicyProvider) *NSNetworkPolicyController {

	controller := &NSNetworkPolicyController{
		client:                  client,
		ksclient:                ksclient,
		informer:                nsnpInformer,
		informerSynced:          nsnpInformer.Informer().HasSynced,
		serviceInformer:         serviceInformer,
		serviceInformerSynced:   serviceInformer.Informer().HasSynced,
		workspaceInformer:       workspaceInformer,
		workspaceInformerSynced: workspaceInformer.Informer().HasSynced,
		namespaceInformer:       namespaceInformer,
		namespaceInformerSynced: namespaceInformer.Informer().HasSynced,
		provider:                policyProvider,
		nsQueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespace"),
		nsnpQueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespacenp"),
	}

	workspaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addWorkspace,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*workspacev1alpha1.Workspace)
			new := oldObj.(*workspacev1alpha1.Workspace)
			if old.Spec.NetworkIsolation == new.Spec.NetworkIsolation {
				return
			}
			controller.addWorkspace(newObj)
		},
	})

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addNamespace,
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			old := oldObj.(*corev1.Namespace)
			new := oldObj.(*corev1.Namespace)
			if old.Annotations[NamespaceNPAnnotationKey] == new.Annotations[NamespaceNPAnnotationKey] {
				return
			}
			controller.addNamespace(newObj)
		},
	})

	nsnpInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(4).Infof("Got ADD event for NSNSP: %#v", obj)
			controller.nsnpEnqueue(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			klog.V(4).Info("Got UPDATE event for NSNSP.")
			klog.V(4).Infof("Old object: \n%#v\n", oldObj)
			klog.V(4).Infof("New object: \n%#v\n", newObj)
			controller.nsnpEnqueue(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(4).Infof("Got DELETE event for NSNP: %#v", obj)
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
	if ok := cache.WaitForCacheSync(stopCh, c.informerSynced, c.serviceInformerSynced, c.workspaceInformerSynced, c.namespaceInformerSynced); !ok {
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
