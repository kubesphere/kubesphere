/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"context"
	"strings"

	"github.com/emicklei/go-restful/v3"
	jsonpatch "github.com/evanphx/json-patch/v5"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kubesphere.io/api/gateway/v1alpha2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
)

const (
	MasterLabel       = "node-role.kubernetes.io/control-plane"
	SvcNameAnnotation = "gateway.kubesphere.io/service-name"
)

type handler struct {
	cache runtimeclient.Reader
}

func (h *handler) ListIngressClassScopes(req *restful.Request, resp *restful.Response) {
	currentNs := req.PathParameter("namespace")
	ctx := req.Request.Context()

	ingressClassScopeList := v1alpha2.IngressClassScopeList{}
	err := h.cache.List(ctx, &ingressClassScopeList)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var ret []v1alpha2.IngressClassScope
	for _, item := range ingressClassScopeList.Items {
		namespaces := item.Spec.Scope.Namespaces
		nsSelector := item.Spec.Scope.NamespaceSelector

		// Specify all namespace
		if len(namespaces) == 0 && nsSelector == "" {
			_ = h.setStatus(ctx, &item)
			ret = append(ret, item)
			continue
		}

		// Specify namespaces
		if len(namespaces) > 0 {
			for _, n := range namespaces {
				if n == currentNs {
					_ = h.setStatus(ctx, &item)
					ret = append(ret, item)
					break
				}
			}
			continue
		}

		// Specify namespaceSelector
		if nsSelector != "" {
			nsList := corev1.NamespaceList{}
			_ = h.cache.List(ctx, &nsList, &runtimeclient.ListOptions{LabelSelector: Selector(nsSelector)})
			for _, n := range nsList.Items {
				if n.Name == currentNs {
					_ = h.setStatus(ctx, &item)
					ret = append(ret, item)
					break
				}
			}
		}
	}

	resp.WriteEntity(ret)
}

func Selector(s string) labels.Selector {
	if selector, err := labels.Parse(s); err != nil {
		return labels.Everything()
	} else {
		return selector
	}
}

func (h *handler) getMasterNodeIp(ctx context.Context) []string {
	internalIps := []string{}
	masters := &corev1.NodeList{}
	err := h.cache.List(ctx, masters, &runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			labels.Set{
				MasterLabel: "",
			})})

	if err != nil {
		klog.Info(err)
		return internalIps
	}

	for _, node := range masters.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				internalIps = append(internalIps, address.Address)
			}
		}
	}
	return internalIps
}

func (h *handler) setStatus(ctx context.Context, ics *v1alpha2.IngressClassScope) (e error) {
	svcKeyStr, exists := ics.Annotations[SvcNameAnnotation]
	if !exists {
		klog.Errorf("Name: %s, Annotation %s not found", ics.Name, SvcNameAnnotation)
		return nil
	}

	svcKeyParts := strings.SplitN(svcKeyStr, "/", 2)
	if len(svcKeyParts) != 2 {
		klog.Errorf("Name: %s, Invalid %s annotation, should follow the namespace/name format", ics.Name, SvcNameAnnotation)
		return nil
	}

	svc := corev1.Service{}
	key := types.NamespacedName{
		Namespace: svcKeyParts[0],
		Name:      svcKeyParts[1],
	}

	if err := h.cache.Get(ctx, key, &svc); err != nil {
		klog.Errorf("Failed to fetch svc %s, %v", key.String(), err)
		return err
	}

	// append selected node ip as loadBalancer ingress ip
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) == 0 {
		rips := h.getMasterNodeIp(ctx)
		for _, rip := range rips {
			gIngress := corev1.LoadBalancerIngress{
				IP: rip,
			}
			svc.Status.LoadBalancer.Ingress = append(svc.Status.LoadBalancer.Ingress, gIngress)
		}
	}

	status := unstructured.Unstructured{
		Object: map[string]interface{}{
			"loadBalancer": svc.Status.LoadBalancer,
			"service":      svc.Spec.Ports,
		},
	}

	target, err := status.MarshalJSON()
	if err != nil {
		return err
	}

	if ics.Status.Raw != nil {
		//merge with origin status
		patch, err := jsonpatch.CreateMergePatch([]byte(`{}`), target)
		if err != nil {
			return err
		}
		modified, err := jsonpatch.MergePatch(ics.Status.Raw, patch)
		if err != nil {
			return err
		}
		ics.Status.Raw = modified
	}
	ics.Status.Raw = target
	return nil
}
