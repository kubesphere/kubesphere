/*
Copyright 2020 The KubeSphere Authors.

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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClientCertificate struct {
	// The client cert file for the targets.
	Cert *SecretKeySelector `json:"cert,omitempty"`
	// The client key file for the targets.
	Key *SecretKeySelector `json:"key,omitempty"`
}

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// RootCA defines the root certificate authorities
	// that clients use when verifying server certificates.
	RootCA *SecretKeySelector `json:"rootCA,omitempty"`
	// The certificate of the client.
	*ClientCertificate `json:"clientCertificate,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `json:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// BasicAuth contains basic HTTP authentication credentials.
type BasicAuth struct {
	Username string             `json:"username"`
	Password *SecretKeySelector `json:"password,omitempty"`
}

// HTTPClientConfig configures an HTTP client.
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// The bearer token for the targets.
	BearerToken *SecretKeySelector `json:"bearerToken,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
}

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// `namespace` is the namespace of the service.
	// Required
	Namespace string `json:"namespace"`

	// `name` is the name of the service.
	// Required
	Name string `json:"name"`

	// `path` is an optional URL path which will be sent in any request to
	// this service.
	// +optional
	Path *string `json:"path,omitempty"`

	// If specified, the port on the service that hosting webhook.
	// Default to 443 for backward compatibility.
	// `port` should be a valid port number (1-65535, inclusive).
	// +optional
	Port *int32 `json:"port,omitempty"`

	// Http scheme, default is http.
	// +optional
	Scheme *string `json:"scheme,omitempty"`
}

// WebhookReceiverSpec defines the desired state of WebhookReceiver
type WebhookReceiverSpec struct {
	// WebhookConfig to be selected for this receiver
	WebhookConfigSelector *metav1.LabelSelector `json:"webhookConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// `url` gives the location of the webhook, in standard URL form
	// (`scheme://host:port/path`). Exactly one of `url` or `service`
	// must be specified.
	//
	// The `host` should not refer to a service running in the cluster; use
	// the `service` field instead. The host might be resolved via external
	// DNS in some api servers (e.g., `kube-apiserver` cannot resolve
	// in-cluster DNS as that would be a layering violation). `host` may
	// also be an IP address.
	//
	// Please note that using `localhost` or `127.0.0.1` as a `host` is
	// risky unless you take great care to run this webhook on all hosts
	// which run an apiserver which might need to make calls to this
	// webhook. Such installs are likely to be non-portable, i.e., not easy
	// to turn up in a new cluster.
	//
	// A path is optional, and if present may be any string permissible in
	// a URL. You may use the path to pass an arbitrary string to the
	// webhook, for example, a cluster identifier.
	//
	// Attempting to use a user or basic auth e.g. "user:password@" is not
	// allowed. Fragments ("#...") and query parameters ("?...") are not
	// allowed, either.
	//
	// +optional
	URL *string `json:"url,omitempty"`

	// `service` is a reference to the service for this webhook. Either
	// `service` or `url` must be specified.
	//
	// If the webhook is running within the cluster, then you should use `service`.
	//
	// +optional
	Service *ServiceReference `json:"service,omitempty"`

	HTTPConfig *HTTPClientConfig `json:"httpConfig,omitempty"`
}

// WebhookReceiverStatus defines the observed state of WebhookReceiver
type WebhookReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=wr
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// WebhookReceiver is the Schema for the webhookreceivers API
type WebhookReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookReceiverSpec   `json:"spec,omitempty"`
	Status WebhookReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebhookReceiverList contains a list of WebhookReceiver
type WebhookReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebhookReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebhookReceiver{}, &WebhookReceiverList{})
}
