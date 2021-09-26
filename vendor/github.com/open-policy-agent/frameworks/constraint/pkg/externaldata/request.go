package externaldata

// RegoRequest is the request for external_data rego function.
type RegoRequest struct {
	// ProviderName is the name of the external data provider.
	ProviderName string `json:"provider"`
	// Keys is the list of keys to send to the external data provider.
	Keys []interface{} `json:"keys"`
}

// ProviderRequest is the API request for the external data provider.
type ProviderRequest struct {
	// APIVersion is the API version of the external data provider.
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind is kind of the external data provider API call. This can be "ProviderRequest" or "ProviderResponse".
	Kind ProviderKind `json:"kind,omitempty"`
	// Request contains the request for the external data provider.
	Request Request `json:"request,omitempty"`
}

// Request is the struct that contains the keys to query.
type Request struct {
	// Keys is the list of keys to send to the external data provider.
	Keys []interface{} `json:"keys,omitempty"`
}

// NewRequest creates a new request for the external data provider.
func NewProviderRequest(keys []interface{}) *ProviderRequest {
	return &ProviderRequest{
		APIVersion: "externaldata.gatekeeper.sh/v1alpha1",
		Kind:       "ProviderRequest",
		Request: Request{
			Keys: keys,
		},
	}
}

// +kubebuilder:validation:Enum=ProviderRequestKind;ProviderResponseKind
type ProviderKind string

const (
	// ProviderRequestKind is the kind of the request.
	ProviderRequestKind ProviderKind = "ProviderRequest"
	// ProviderResponseKind is the kind of the response.
	ProviderResponseKind ProviderKind = "ProviderResponse"
)
