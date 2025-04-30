/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type APIServiceSpec struct {
	Group    string `json:"group,omitempty"`
	Version  string `json:"version,omitempty"`
	Endpoint `json:",inline"`
}

type APIServiceStatus struct {
	State      string             `json:"state,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// APIService is a special resource used in Ks-apiserver
// declares a directional proxy path for a resource type API，
// it's similar to Kubernetes API Aggregation Layer.
type APIService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServiceSpec   `json:"spec,omitempty"`
	Status APIServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type APIServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIService `json:"items"`
}

type JSBundleSpec struct {
	// +optional
	Raw []byte `json:"raw,omitempty"`
	// +optional
	RawFrom RawFrom `json:"rawFrom,omitempty"`
	// +optional
	Assets Assets `json:"assets,omitempty"`
}

type Assets struct {
	Style *AuxiliaryStyle `json:"style,omitempty"`
	Files []FileLocation  `json:"files,omitempty"`
}

type AuxiliaryStyle struct {
	Link     string `json:"link,omitempty"`
	Endpoint `json:",inline"`
}

type FileLocation struct {
	Name string `json:"name,omitempty"`
	Link string `json:"link,omitempty"`
	// Set the MIME Type of the file, if not specified, it will be provided by the content-type response header in the upstream service by default.
	// +optional
	MIMEType *string `json:"mimeType,omitempty"`
	Endpoint `json:",inline"`
}

type RawFrom struct {
	Endpoint `json:",inline"`
	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *ConfigMapKeyRef `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty"`
}

type ConfigMapKeyRef struct {
	corev1.ConfigMapKeySelector `json:",inline"`
	Namespace                   string `json:"namespace"`
}

type SecretKeyRef struct {
	corev1.SecretKeySelector `json:",inline"`
	Namespace                string `json:"namespace"`
}

type JSBundleStatus struct {
	// Link is the path for downloading JS file, default to "/dist/{jsBundleName}/index.js".
	// +optional
	Link  string `json:"link,omitempty"`
	State string `json:"state,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// JSBundle declares a js bundle that needs to be injected into ks-console,
// the endpoint can be provided by a service or a static file.
type JSBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JSBundleSpec   `json:"spec,omitempty"`
	Status JSBundleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type JSBundleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JSBundle `json:"items"`
}

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

type ExtensionEntrySpec struct {
	Entries []runtime.RawExtension `json:"entries,omitempty"`
}

type ExtensionEntryStatus struct {
	State string `json:"state,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// ExtensionEntry declares an entry endpoint that needs to be injected into ks-console.
type ExtensionEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtensionEntrySpec   `json:"spec,omitempty"`
	Status ExtensionEntryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ExtensionEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtensionEntry `json:"items"`
}
