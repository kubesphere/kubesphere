package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	azureIMDSEndpoint = "http://169.254.169.254/metadata/identity/oauth2/token"
	defaultAPIVersion = "2018-02-01"
	defaultResource   = "https://storage.azure.com/"
	timeout           = 5 * time.Second
)

// azureManagedIdentitiesToken holds a token for managed identities for Azure resources
type azureManagedIdentitiesToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
	ExpiresOn   string `json:"expires_on"`
	NotBefore   string `json:"not_before"`
	Resource    string `json:"resource"`
	TokenType   string `json:"token_type"`
}

// azureManagedIdentitiesError represents an error fetching an azureManagedIdentitiesToken
type azureManagedIdentitiesError struct {
	Err         string `json:"error"`
	Description string `json:"error_description"`
	Endpoint    string
	StatusCode  int
}

func (e *azureManagedIdentitiesError) Error() string {
	return fmt.Sprintf("%v %s retrieving azure token from %s: %s", e.StatusCode, e.Err, e.Endpoint, e.Description)
}

// azureManagedIdentitiesAuthPlugin uses an azureManagedIdentitiesToken.AccessToken for bearer authorization
type azureManagedIdentitiesAuthPlugin struct {
	Endpoint   string `json:"endpoint"`
	APIVersion string `json:"api_version"`
	Resource   string `json:"resource"`
	ObjectID   string `json:"object_id"`
	ClientID   string `json:"client_id"`
	MiResID    string `json:"mi_res_id"`
}

func (ap *azureManagedIdentitiesAuthPlugin) NewClient(c Config) (*http.Client, error) {
	if c.Type == "oci" {
		return nil, errors.New("azure managed identities auth: OCI service not supported")
	}

	if ap.Endpoint == "" {
		ap.Endpoint = azureIMDSEndpoint
	}

	if ap.Resource == "" {
		ap.Resource = defaultResource
	}

	if ap.APIVersion == "" {
		ap.APIVersion = defaultAPIVersion
	}

	t, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}

	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

func (ap *azureManagedIdentitiesAuthPlugin) Prepare(req *http.Request) error {
	token, err := azureManagedIdentitiesTokenRequest(
		ap.Endpoint, ap.APIVersion, ap.Resource,
		ap.ObjectID, ap.ClientID, ap.MiResID,
	)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	return nil
}

// azureManagedIdentitiesTokenRequest fetches an azureManagedIdentitiesToken
func azureManagedIdentitiesTokenRequest(
	endpoint, apiVersion, resource, objectID, clientID, miResID string,
) (azureManagedIdentitiesToken, error) {
	var token azureManagedIdentitiesToken
	e := buildAzureManagedIdentitiesRequestPath(endpoint, apiVersion, resource, objectID, clientID, miResID)

	request, err := http.NewRequest("GET", e, nil)
	if err != nil {
		return token, err
	}
	request.Header.Add("Metadata", "true")

	httpClient := http.Client{Timeout: timeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return token, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return token, err
	}

	if s := response.StatusCode; s != http.StatusOK {
		var azureError azureManagedIdentitiesError
		err = json.Unmarshal(data, &azureError)
		if err != nil {
			return token, err
		}

		azureError.Endpoint = e
		azureError.StatusCode = s
		return token, &azureError
	}

	err = json.Unmarshal(data, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

// buildAzureManagedIdentitiesRequestPath constructs the request URL for an Azure managed identities token request
func buildAzureManagedIdentitiesRequestPath(
	endpoint, apiVersion, resource, objectID, clientID, miResID string,
) string {
	params := url.Values{
		"api-version": []string{apiVersion},
		"resource":    []string{resource},
	}

	if objectID != "" {
		params.Add("object_id", objectID)
	}

	if clientID != "" {
		params.Add("client_id", clientID)
	}

	if miResID != "" {
		params.Add("mi_res_id", miResID)
	}

	return endpoint + "?" + params.Encode()
}
