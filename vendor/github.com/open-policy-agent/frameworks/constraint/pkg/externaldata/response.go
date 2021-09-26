package externaldata

import (
	"bytes"
	"encoding/json"

	"github.com/open-policy-agent/opa/ast"
)

// RegoResponse is the response inside rego.
type RegoResponse struct {
	// Responses contains the response from the provider.
	// In each element of the outer array, the first element is the key and the second is the corresponding value from the provider.
	Responses [][]interface{} `json:"responses"`
	// Errors contains the errors from the provider.
	// In each item of the outer array, the first element is the key and the second is the corresponding error from the provider.
	Errors [][]interface{} `json:"errors"`
	// StatusCode contains the status code of the response.
	StatusCode int `json:"status_code"`
	// SystemError is the system error of the response.
	SystemError string `json:"system_error"`
}

// ProviderResponse is the API response from a provider.
type ProviderResponse struct {
	// APIVersion is the API version of the external data provider.
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind is kind of the external data provider API call. This can be "ProviderRequest" or "ProviderResponse".
	Kind ProviderKind `json:"kind,omitempty"`
	// Response contains the response from the provider.
	Response Response `json:"response,omitempty"`
}

// Response is the struct that holds the response from a provider.
type Response struct {
	// Idempotent indicates that the responses from the provider are idempotent.
	// Applies to mutation only and must be true for mutation.
	Idempotent bool `json:"idempotent,omitempty"`
	// Items contains the key, value and error from the provider.
	Items []Item `json:"items,omitempty"`
	// SystemError is the system error of the response.
	SystemError string `json:"systemError,omitempty"`
}

// Items is the struct that contains the key, value or error from a provider response.
type Item struct {
	// Key is the request from the provider.
	Key interface{} `json:"key,omitempty"`
	// Value is the response from the provider.
	Value interface{} `json:"value,omitempty"`
	// Error is the error from the provider.
	Error string `json:"error,omitempty"`
}

// NewRegoResponse creates a new rego response from the given provider response.
func NewRegoResponse(statusCode int, pr *ProviderResponse) *RegoResponse {
	responses := make([][]interface{}, 0)
	errors := make([][]interface{}, 0)

	for _, item := range pr.Response.Items {
		if item.Error != "" {
			errors = append(errors, []interface{}{item.Key, item.Error})
		} else {
			responses = append(responses, []interface{}{item.Key, item.Value})
		}
	}

	return &RegoResponse{
		Responses:   responses,
		Errors:      errors,
		StatusCode:  statusCode,
		SystemError: pr.Response.SystemError,
	}
}

func PrepareRegoResponse(regoResponse *RegoResponse) (*ast.Term, error) {
	rr, err := json.Marshal(regoResponse)
	if err != nil {
		return nil, err
	}
	v, err := ast.ValueFromReader(bytes.NewReader(rr))
	if err != nil {
		return nil, err
	}
	return ast.NewTerm(v), nil
}

func HandleError(statusCode int, err error) (*ast.Term, error) {
	regoResponse := RegoResponse{
		StatusCode:  statusCode,
		SystemError: err.Error(),
	}
	return PrepareRegoResponse(&regoResponse)
}
