package v1alpha2

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"time"
)

func (h handler) intersect(nsFilter []string, nsSearch []string, wsFilter []string, wsSearch []string) map[string]time.Time {
	nsList, err := h.k.Kubernetes().CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list namespace, error: %s", err)
		return nil
	}

	inner := make(map[string]time.Time)

	// if no search condition is set on both namespace and workspace,
	// then return all namespaces
	if nsSearch == nil && nsFilter == nil && wsSearch == nil && wsFilter == nil {
		for _, ns := range nsList.Items {
			inner[ns.Name] = ns.CreationTimestamp.Time
		}
	} else {
		for _, ns := range nsList.Items {
			if stringutils.StringIn(ns.Name, nsFilter) ||
				stringutils.StringIn(ns.Annotations[constants.WorkspaceLabelKey], wsFilter) ||
				containsIn(ns.Name, nsSearch) ||
				containsIn(ns.Annotations[constants.WorkspaceLabelKey], wsSearch) {
				inner[ns.Name] = ns.CreationTimestamp.Time
			}
		}
	}

	return inner
}

func containsIn(str string, subStrs []string) bool {
	for _, sub := range subStrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

func (h handler) withCreationTime(name string) map[string]time.Time {
	ns, err := h.k.Kubernetes().CoreV1().Namespaces().Get(name, v1.GetOptions{})
	if err == nil {
		return map[string]time.Time{name: ns.CreationTimestamp.Time}
	}
	return nil
}
