/*
Copyright 2022 The KubeSphere Authors.

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
	// Performs substring replacements on the URI.
	URISubstring []string `json:"uriSubstring,omitempty"`
	// Performs regular expression replacements on the URI path.
	PathRegexp []string `json:"pathRegexp,omitempty"`
	// Sets, adds (with the + prefix), deletes (with the - prefix), or performs a replacement (by using two arguments, a search and replacement) in a request header going upstream to the backend.
	HeaderUp []string `json:"headerUp,omitempty"`
	// Sets, adds (with the + prefix), deletes (with the - prefix), or performs a replacement (by using two arguments, a search and replacement) in a response header coming downstream from the backend.
	HeaderDown []string `json:"headerDown,omitempty"`
	// Change Host header for name-based virtual hosted sites.
	ChangeOrigin bool `json:"changeOrigin,omitempty"`
	// InterceptRedirects determines whether the proxy should sniff backend responses for redirects, only allows redirects to the same host.
	InterceptRedirects bool `json:"interceptRedirects,omitempty"`
	//  WrapTransport indicates whether the provided Transport should be wrapped with default proxy transport behavior (URL rewriting, X-Forwarded-* header setting)
	WrapTransport bool `json:"wrapTransport,omitempty"`
	// Add auth proxy header to requests
	AuthProxy bool `json:"authProxy,omitempty"`
}

type ReverseProxyStatus struct {
	State string `json:"state,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="extensions",scope="Cluster"

// ReverseProxy defines the rules for reverse proxies,
// it's useful to exposing APIs to ks-console or other extensions.
type ReverseProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReverseProxySpec   `json:"spec,omitempty"`
	Status ReverseProxyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReverseProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReverseProxy `json:"items"`
}
