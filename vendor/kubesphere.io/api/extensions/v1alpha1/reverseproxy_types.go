package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Matcher struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type ReverseProxySpec struct {
	Matcher    Matcher    `json:"matcher,omitempty"`
	Upstream   Endpoint   `json:"upstream,omitempty"`
	Directives Directives `json:"directives,omitempty"`
}

type Directives struct {
	// Changes the request's HTTP verb.
	Method string `json:"method,omitempty"`
	// Strips the given prefix from the beginning of the URI path.
	StripPathPrefix string `json:"stripPathPrefix,omitempty"`
	// Strips the given suffix from the end of the URI path.
	StripPathSuffix string `json:"stripPathSuffix,omitempty"`
	// Sets, adds (with the + prefix), deletes (with the - prefix), or performs a replacement (by using two arguments, a search and replacement) in a request header going upstream to the backend.
	HeaderUp []string `json:"headerUp,omitempty"`
	// Sets, adds (with the + prefix), deletes (with the - prefix), or performs a replacement (by using two arguments, a search and replacement) in a response header coming downstream from the backend.
	HeaderDown []string `json:"headerDown,omitempty"`
	// Reject to forward redirect response
	RejectForwardingRedirects bool `json:"rejectForwardingRedirects,omitempty"`
	//  WrapTransport indicates whether the provided Transport should be wrapped with default proxy transport behavior (URL rewriting, X-Forwarded-* header setting)
	WrapTransport bool `json:"wrapTransport,omitempty"`
	// Add auth proxy header to requests
	AuthProxy  bool     `json:"authProxy,omitempty"`
	Rewrite    []string `json:"rewrite,omitempty"`
	Replace    []string `json:"replace,omitempty"`
	PathRegexp []string `json:"pathRegexp,omitempty"`
}

type ReverseProxyStatus struct {
	State string `json:"state,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

type ReverseProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReverseProxySpec   `json:"spec,omitempty"`
	Status ReverseProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ReverseProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReverseProxy `json:"items"`
}
