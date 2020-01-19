package logging

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"time"
)


// list namespaces matching the search condition
func ListMatchedNamespaces(informers informers.SharedInformerFactory, nsFilter []string, nsSearch []string, wsFilter []string, wsSearch []string) (bool, []string) {

	nsLister := informers.Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to list namespace, error: %s", err)
		return true, nil
	}

	var namespaces []string

	// if no search condition is set on both namespace and workspace,
	// then return all namespaces
	if nsSearch == nil && nsFilter == nil && wsSearch == nil && wsFilter == nil {
		for _, ns := range nsList {
			namespaces = append(namespaces, ns.Name)
		}
		return false, namespaces
	}

	for _, ns := range nsList {
		if stringutils.StringIn(ns.Name, nsFilter) ||
			stringutils.StringIn(ns.Annotations[constants.WorkspaceLabelKey], wsFilter) ||
			containsIn(ns.Name, nsSearch) ||
			containsIn(ns.Annotations[constants.WorkspaceLabelKey], wsSearch) {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// if namespaces is equal to nil, indicates no namespace matched
	// it causes the query to return no result
	return namespaces == nil, namespaces
}

func containsIn(str string, subStrs []string) bool {
	for _, sub := range subStrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

func WithCreationTimestamp(informers informers.SharedInformerFactory, namespaces []string) map[string]time.Time {

	namespaceWithCreationTime := make(map[string]time.Time)

	nsLister := informers.Core().V1().Namespaces().Lister()
	for _, item := range namespaces {
		ns, err := nsLister.Get(item)
		if err != nil {
			// the ns doesn't exist
			continue
		}
		namespaceWithCreationTime[ns.Name] = ns.CreationTimestamp.Time
	}

	return namespaceWithCreationTime
}
