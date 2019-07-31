package wsnetworkpolicy

import (
	"fmt"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	ks8network "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	wsnpapi "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
)

const (
	workspaceSelectorLabel      = "kubesphere.io/workspace"
	workspaceNetworkPolicyLabel = "networking.kubesphere.io/wsnp"

	MessageResourceExists = "Resource %q already exists and is not managed by WorkspaceNetworkPolicy"
	ErrResourceExists     = "ErrResourceExists"
)

var everything = labels.Everything()
var reconcileCount = 0

// NetworkPolicyNameForWSNP return the name of the networkpolicy owned by this WNSP
func NetworkPolicyNameForWSNP(wsnp string) string {
	return wsnp + "-np"
}

func (c *controller) reconcile(key string) error {
	reconcileCount++
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	olog := log.WithName(name)
	olog.Info("Begin to reconcile")
	owner, err := c.wsnpLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("WSNP '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	namespaces, err := c.listNamespacesInWorkspace(owner.Spec.Workspace)
	if err != nil {
		return err
	}
	var errs []error
	for _, ns := range namespaces {
		err = c.reconcileNamespace(ns.Name, owner)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errutil.NewAggregate(errs)
}

func (c *controller) reconcileNamespace(name string, wsnp *wsnpapi.WorkspaceNetworkPolicy) error {
	npname := NetworkPolicyNameForWSNP(wsnp.Name)
	np, err := c.generateNPForNamesapce(name, wsnp)
	if err != nil {
		log.Error(nil, "Failed to generate NetworkPolicy", "wsnp", wsnp, "namespace", name)
		return err
	}
	old, err := c.networkPolicyLister.NetworkPolicies(name).Get(npname)
	if errors.IsNotFound(err) {
		_, err = c.kubeClientset.NetworkingV1().NetworkPolicies(name).Create(np)
		if err != nil {
			log.Error(err, "cannot create networkpolicy of this wsnp", wsnp)
			return err
		}
		return nil
	}
	if err != nil {
		log.Error(err, "Failed to get networkPolicy")
		return err
	}
	if !metav1.IsControlledBy(old, wsnp) {
		msg := fmt.Sprintf(MessageResourceExists, old.Name)
		c.recorder.Event(wsnp, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !reflect.DeepEqual(old.Spec, np.Spec) {
		log.V(2).Info("Detect network policy changed, updating network policy", "the old one", old.Spec, "the new one", np.Spec)
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			_, err = c.kubeClientset.NetworkingV1().NetworkPolicies(name).Update(np)
			return err
		})
		if err != nil {
			log.Error(err, "Failed to update wsnp")
			return err
		}
		log.V(2).Info("updating completed")
	}
	return nil
}

func (c *controller) generateNPForNamesapce(ns string, wsnp *wsnpapi.WorkspaceNetworkPolicy) (*ks8network.NetworkPolicy, error) {
	np := &ks8network.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyNameForWSNP(wsnp.Name),
			Namespace: ns,
			Labels:    map[string]string{workspaceNetworkPolicyLabel: wsnp.Name},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(wsnp, wsnpapi.SchemeGroupVersion.WithKind("WorkspaceNetworkPolicy")),
			},
		},
		Spec: ks8network.NetworkPolicySpec{
			PolicyTypes: wsnp.Spec.PolicyTypes,
		},
	}

	if wsnp.Spec.Ingress != nil {
		np.Spec.Ingress = make([]ks8network.NetworkPolicyIngressRule, len(wsnp.Spec.Ingress))
		for index, ing := range wsnp.Spec.Ingress {
			ingRule, err := c.transformWSNPIngressToK8sIngress(ing)
			if err != nil {
				return nil, err
			}
			np.Spec.Ingress[index] = *ingRule
		}
	}
	return np, nil
}

func (c *controller) transformWSNPIngressToK8sIngress(rule wsnpapi.WorkspaceNetworkPolicyIngressRule) (*ks8network.NetworkPolicyIngressRule, error) {
	k8srule := &ks8network.NetworkPolicyIngressRule{
		Ports: rule.Ports,
		From:  make([]ks8network.NetworkPolicyPeer, len(rule.From)),
	}
	for index, f := range rule.From {
		k8srule.From[index] = f.NetworkPolicyPeer
		if f.WorkspaceSelector != nil {
			if f.WorkspaceSelector.Size() == 0 {
				k8srule.From[index].NamespaceSelector = &metav1.LabelSelector{}
			} else {
				selector, err := metav1.LabelSelectorAsSelector(f.WorkspaceSelector)
				if err != nil {
					log.Error(err, "Failed to convert label selectors")
					return nil, err
				}
				ws, err := c.workspaceLister.List(selector)
				if err != nil {
					log.Error(err, "Failed to list workspaces")
					return nil, err
				}
				if len(ws) == 0 {
					log.Info("ws selector doesnot match anything")
					continue
				}
				if k8srule.From[index].NamespaceSelector == nil {
					k8srule.From[index].NamespaceSelector = &metav1.LabelSelector{}
				}
				if len(ws) == 1 {
					if k8srule.From[index].NamespaceSelector.MatchLabels == nil {
						k8srule.From[index].NamespaceSelector.MatchLabels = make(map[string]string)
					}
					k8srule.From[index].NamespaceSelector.MatchLabels[workspaceSelectorLabel] = ws[0].Name
				} else {
					if k8srule.From[index].NamespaceSelector.MatchExpressions == nil {
						k8srule.From[index].NamespaceSelector.MatchExpressions = make([]metav1.LabelSelectorRequirement, 0)
					}
					re := metav1.LabelSelectorRequirement{
						Key:      workspaceSelectorLabel,
						Operator: metav1.LabelSelectorOpIn,
						Values:   make([]string, len(ws)),
					}
					for index, w := range ws {
						re.Values[index] = w.Name
					}
					sort.Strings(re.Values)
					k8srule.From[index].NamespaceSelector.MatchExpressions = append(k8srule.From[index].NamespaceSelector.MatchExpressions, re)
				}
			}
		}
	}
	return k8srule, nil
}
func (c *controller) listNamespacesInWorkspace(workspace string) ([]*corev1.Namespace, error) {
	selector, err := labels.Parse(workspaceSelectorLabel + "==" + workspace)
	if err != nil {
		log.Error(err, "Failed to parse label selector")
		return nil, err
	}
	namespaces, err := c.namespaceLister.List(selector)
	if err != nil {
		log.Error(err, "Failed to list namespaces in this workspace")
		return nil, err
	}
	return namespaces, nil
}
