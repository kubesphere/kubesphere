package v1alpha1

import "fmt"

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// namespace is the namespace of the service.
	// Required
	Namespace string `json:"namespace"`
	// name is the name of the service.
	// Required
	Name string `json:"name"`

	// path is an optional URL path at which the upstream will be contacted.
	// +optional
	Path *string `json:"path,omitempty"`

	// port is an optional service port at which the upstream will be contacted.
	// `port` should be a valid port number (1-65535, inclusive).
	// Defaults to 443 for backward compatibility.
	// +optional
	Port *int32 `json:"port,omitempty"`
}

type Endpoint struct {
	// `url` gives the location of the upstream, in standard URL form
	// (`scheme://host:port/path`). Exactly one of `url` or `service`
	// must be specified.
	// +optional
	URL *string `json:"url,omitempty"`
	// service is a reference to the service for this endpoint. Either
	// service or url must be specified.
	// the scheme is default to HTTPS.
	// +optional
	Service *ServiceReference `json:"service,omitempty"`
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

func (in *Endpoint) RawURL() string {
	var rawURL string
	if in.URL != nil {
		rawURL = *in.URL
	} else if in.Service != nil {
		var port int32 = 443
		var path = ""
		if in.Service.Port != nil {
			port = *in.Service.Port
		}
		if in.Service.Path != nil {
			path = *in.Service.Path
		}
		rawURL = fmt.Sprintf("https://%s.%s.svc:%d%s",
			in.Service.Name,
			in.Service.Namespace,
			port, path)
	}
	return rawURL
}
