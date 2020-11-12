package security

import (
	"encoding/base64"
	"fmt"
)

// Identity security details about a client.
type Identity struct {
	CertFile       string `yaml:"cert_file"`
	PrivateKeyFile string `yaml:"private_key_file"`
}

// Credentials provides information when needing to authenticate to remote endpoints.
// Credentials are either a username/password or a bearer token, but not both.
type Credentials struct {
	Username       string `yaml:",omitempty"`
	Password       string `yaml:",omitempty"`
	Token          string `yaml:",omitempty"`
	AllowAnonymous bool   `yaml:"allow_anonymous,omitempty"`
}

// TLS options - SkipCertificateValidation will disable server certificate verification - the client
// will accept any certificate presented by the server and any host name in that certificate.
type TLS struct {
	SkipCertificateValidation bool `yaml:"skip_certificate_validation,omitempty"`
}

// ValidateCredentials makes sure that if username is provided, so is password (and vice versa)
// and also makes sure if username/password is provided that token is not (and vice versa).
// It is valid to have nothing defined (no username, password, nor token), but if nothing is
// defined and the "AllowAnonymous" flag is false, this usually means the person who
// installed Kiali most likely forgot to set credentials - therefore access should always be denied.
// If nothing is defined and the "AllowAnonymous" flag is true, this means anonymous access is specifically allowed.
// If the "AllowAnonymous" flag is true but non-empty credentials are defined, an error results.
func (c *Credentials) ValidateCredentials() error {
	if c.Username != "" && c.Password == "" {
		return fmt.Errorf("A password must be provided if a username is set")
	}

	if c.Username == "" && c.Password != "" {
		return fmt.Errorf("A username must be provided if a password is set")
	}

	if c.Username != "" && c.Token != "" {
		return fmt.Errorf("Username/password cannot be specified if a token is specified also. Only Username/Password or Token can be set but not both")
	}

	if c.AllowAnonymous && (c.Username != "" || c.Token != "") {
		return fmt.Errorf("The 'AllowAnonymous' flag is true but non-empty credentials exist")
	}

	return nil
}

// GetHTTPAuthHeader provides the authentication ehader name and value (can be empty), or an error
func (c *Credentials) GetHTTPAuthHeader() (headerName string, headerValue string, err error) {
	// if no credentials are provided, this is fine, we are just going to do an insecure request
	if c == nil {
		return "", "", nil
	}

	if err := c.ValidateCredentials(); err != nil {
		return "", "", err
	}

	if c.Token != "" {
		headerName = "Authorization"
		headerValue = fmt.Sprintf("Bearer %s", c.Token)
	} else if c.Username != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password)))
		headerName = "Authorization"
		headerValue = fmt.Sprintf("Basic %s", creds)
	} else {
		headerName = ""
		headerValue = ""
	}

	return headerName, headerValue, nil
}
