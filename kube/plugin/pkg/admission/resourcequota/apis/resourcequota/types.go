package resourcequota

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// Configuration provides configuration for the ResourceQuota admission controller.
type Configuration struct {
	metav1.TypeMeta

	// LimitedResources whose consumption is limited by default.
	// +optional
	LimitedResources []LimitedResource
}

// LimitedResource matches a resource whose consumption is limited by default.
// To consume the resource, there must exist an associated quota that limits
// its consumption.
type LimitedResource struct {

	// APIGroup is the name of the APIGroup that contains the limited resource.
	// +optional
	APIGroup string `json:"apiGroup,omitempty"`

	// Resource is the name of the resource this rule applies to.
	// For example, if the administrator wants to limit consumption
	// of a storage resource associated with persistent volume claims,
	// the value would be "persistentvolumeclaims".
	Resource string `json:"resource"`

	// For each intercepted request, the quota system will evaluate
	// its resource usage.  It will iterate through each resource consumed
	// and if the resource contains any substring in this listing, the
	// quota system will ensure that there is a covering quota.  In the
	// absence of a covering quota, the quota system will deny the request.
	// For example, if an administrator wants to globally enforce that
	// that a quota must exist to consume persistent volume claims associated
	// with any storage class, the list would include
	// ".storageclass.storage.k8s.io/requests.storage"
	MatchContains []string

	// For each intercepted request, the quota system will figure out if the input object
	// satisfies a scope which is present in this listing, then
	// quota system will ensure that there is a covering quota.  In the
	// absence of a covering quota, the quota system will deny the request.
	// For example, if an administrator wants to globally enforce that
	// a quota must exist to create a pod with "cluster-services" priorityclass
	// the list would include
	// "PriorityClassNameIn=cluster-services"
	// +optional
	//	MatchScopes []string `json:"matchScopes,omitempty"`
	MatchScopes []corev1.ScopedResourceSelectorRequirement `json:"matchScopes,omitempty"`
}
