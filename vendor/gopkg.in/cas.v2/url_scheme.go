package cas

import (
	"net/url"
	"path"
)

// URLScheme creates the url which are required to handle the cas protocol.
type URLScheme interface {
	Login() (*url.URL, error)
	Logout() (*url.URL, error)
	Validate() (*url.URL, error)
	ServiceValidate() (*url.URL, error)
	RestGrantingTicket() (*url.URL, error)
	RestServiceTicket(tgt string) (*url.URL, error)
	RestLogout(tgt string) (*url.URL, error)
}

// NewDefaultURLScheme creates a URLScheme which uses the cas default urls
func NewDefaultURLScheme(base *url.URL) *DefaultURLScheme {
	return &DefaultURLScheme{
		base:                base,
		LoginPath:           "login",
		LogoutPath:          "logout",
		ValidatePath:        "validate",
		ServiceValidatePath: "serviceValidate",
		RestEndpoint:        path.Join("v1", "tickets"),
	}
}

// DefaultURLScheme is a configurable URLScheme. Use NewDefaultURLScheme to create DefaultURLScheme with the default cas
// urls.
type DefaultURLScheme struct {
	base                *url.URL
	LoginPath           string
	LogoutPath          string
	ValidatePath        string
	ServiceValidatePath string
	RestEndpoint        string
}

// Login returns the url for the cas login page
func (scheme *DefaultURLScheme) Login() (*url.URL, error) {
	return scheme.createURL(scheme.LoginPath)
}

// Logout returns the url for the cas logut page
func (scheme *DefaultURLScheme) Logout() (*url.URL, error) {
	return scheme.createURL(scheme.LogoutPath)
}

// Validate returns the url for the request validation endpoint
func (scheme *DefaultURLScheme) Validate() (*url.URL, error) {
	return scheme.createURL(scheme.ValidatePath)
}

// ServiceValidate returns the url for the service validation endpoint
func (scheme *DefaultURLScheme) ServiceValidate() (*url.URL, error) {
	return scheme.createURL(scheme.ServiceValidatePath)
}

// RestGrantingTicket returns the url for requesting an granting ticket via rest api
func (scheme *DefaultURLScheme) RestGrantingTicket() (*url.URL, error) {
	return scheme.createURL(scheme.RestEndpoint)
}

// RestServiceTicket returns the url for requesting an service ticket via rest api
func (scheme *DefaultURLScheme) RestServiceTicket(tgt string) (*url.URL, error) {
	return scheme.createURL(path.Join(scheme.RestEndpoint, tgt))
}

// RestLogout returns the url for destroying an granting ticket via rest api
func (scheme *DefaultURLScheme) RestLogout(tgt string) (*url.URL, error) {
	return scheme.createURL(path.Join(scheme.RestEndpoint, tgt))
}

func (scheme *DefaultURLScheme) createURL(urlPath string) (*url.URL, error) {
	return scheme.base.Parse(path.Join(scheme.base.Path, urlPath))
}
