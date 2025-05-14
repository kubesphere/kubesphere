// Copyright 2019 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package rest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/internal/jwx/jwa"
	"github.com/open-policy-agent/opa/internal/jwx/jws"
	"github.com/open-policy-agent/opa/internal/jwx/jws/sign"
	"github.com/open-policy-agent/opa/internal/providers/aws"
	"github.com/open-policy-agent/opa/internal/uuid"
	"github.com/open-policy-agent/opa/v1/keys"
	"github.com/open-policy-agent/opa/v1/logging"
)

const (
	// Default to s3 when the service for sigv4 signing is not specified for backwards compatibility
	awsSigv4SigningDefaultService = "s3"
	// Default to urn:ietf:params:oauth:client-assertion-type:jwt-bearer for ClientAssertionType when not specified
	defaultClientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

// DefaultTLSConfig defines standard TLS configurations based on the Config
func DefaultTLSConfig(c Config) (*tls.Config, error) {
	t := &tls.Config{}
	url, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}
	if url.Scheme == "https" {
		t.InsecureSkipVerify = c.AllowInsecureTLS
	}

	if c.TLS != nil && c.TLS.CACert != "" {
		caCert, err := os.ReadFile(c.TLS.CACert)
		if err != nil {
			return nil, err
		}

		var rootCAs *x509.CertPool
		if c.TLS.SystemCARequired {
			rootCAs, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
		} else {
			rootCAs = x509.NewCertPool()
		}

		ok := rootCAs.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, errors.New("unable to parse and append CA certificate to certificate pool")
		}
		t.RootCAs = rootCAs
	}

	return t, nil
}

// DefaultRoundTripperClient is a reasonable set of defaults for HTTP auth plugins
func DefaultRoundTripperClient(t *tls.Config, timeout int64) *http.Client {
	// Ensure we use a http.Transport with proper settings: the zero values are not
	// a good choice, as they cause leaking connections:
	// https://github.com/golang/go/issues/19620

	// copy, we don't want to alter the default client's Transport
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.ResponseHeaderTimeout = time.Duration(timeout) * time.Second
	tr.TLSClientConfig = t

	c := *http.DefaultClient
	c.Transport = tr
	return &c
}

// defaultAuthPlugin represents baseline 'no auth' behavior if no alternative plugin is specified for a service
type defaultAuthPlugin struct{}

func (*defaultAuthPlugin) NewClient(c Config) (*http.Client, error) {
	t, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}
	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

func (*defaultAuthPlugin) Prepare(*http.Request) error {
	return nil
}

type serverTLSConfig struct {
	CACert           string `json:"ca_cert,omitempty"`
	SystemCARequired bool   `json:"system_ca_required,omitempty"`
}

// bearerAuthPlugin represents authentication via a bearer token in the HTTP Authorization header
type bearerAuthPlugin struct {
	Token     string `json:"token"`
	TokenPath string `json:"token_path"`
	Scheme    string `json:"scheme,omitempty"`

	// encode is set to true for the OCIDownloader because
	// it expects tokens in plain text but needs them in base64.
	encode bool
	logger logging.Logger
}

func (ap *bearerAuthPlugin) NewClient(c Config) (*http.Client, error) {
	t, err := DefaultTLSConfig(c)

	ap.logger = c.logger

	if err != nil {
		return nil, err
	}

	if ap.Token != "" && ap.TokenPath != "" {
		return nil, errors.New("invalid config: specify a value for either the \"token\" or \"token_path\" field")
	}

	if ap.Scheme == "" {
		ap.Scheme = "Bearer"
	}

	if c.Type == "oci" {
		// Standard rest clients use the bearer token as it is defined in the Config
		// but the OCIDownloader needs it encoded to base64 before using to sign a request.
		ap.encode = true
	}

	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

func (ap *bearerAuthPlugin) Prepare(req *http.Request) error {
	token := ap.Token
	if ap.logger == nil {
		ap.logger = logging.Get()
	}

	if ap.TokenPath != "" {
		bytes, err := os.ReadFile(ap.TokenPath)
		if err != nil {
			return err
		}
		token = strings.TrimSpace(string(bytes))
	}

	if ap.encode {
		token = base64.StdEncoding.EncodeToString([]byte(token))
	}

	if req.Response != nil && (req.Response.StatusCode == http.StatusPermanentRedirect || req.Response.StatusCode == http.StatusTemporaryRedirect) {
		ap.logger.Debug("not attaching authorization header as the response contains a redirect")
	} else {
		ap.logger.Debug("attaching authorization header")
		req.Header.Add("Authorization", fmt.Sprintf("%v %v", ap.Scheme, token))
	}
	return nil
}

type tokenEndpointResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type awsKmsKeyConfig struct {
	Name      string `json:"name"`
	Algorithm string `json:"algorithm"`
}

func convertSignatureToBase64(alg string, der []byte) (string, error) {
	r, s, derErr := pointsFromDER(der)
	if derErr != nil {
		return "", fmt.Errorf("failed to read points from der %v", derErr)
	}

	signatureData, err := convertPointsToBase64(alg, r.Bytes(), s.Bytes())
	if err != nil {
		return "", err
	}
	return signatureData, nil
}

func pointsFromDER(der []byte) (R, S *big.Int, err error) { //nolint:gocritic
	R, S = &big.Int{}, &big.Int{}
	data := asn1.RawValue{}
	if _, err := asn1.Unmarshal(der, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshall the signature from DER format %v", err)

	}
	// https://docs.aws.amazon.com/kms/latest/APIReference/API_Sign.html#API_Sign_ResponseSyntax
	// https://datatracker.ietf.org/doc/html/rfc3279#section-2.2.3
	// The format of our DER string is 0x02 + rlen + r + 0x02 + slen + s
	rLen := data.Bytes[1] // The entire length of R + offset of 2 for 0x02 and rlen
	r := data.Bytes[2 : rLen+2]
	// Ignore the next 0x02 and slen bytes and just take the start of S to the end of the byte array
	s := data.Bytes[rLen+4:]
	R.SetBytes(r)
	S.SetBytes(s)
	return
}

func convertPointsToBase64(alg string, r, s []byte) (string, error) {
	curveBits, err := retrieveCurveBits(alg)
	if err != nil {
		return "", err
	}
	keyBytes := curveBits / 8
	if curveBits%8 > 0 {
		keyBytes++
	}
	// We serialize the outputs (r and s) into big-endian byte arrays and pad
	// them with zeros on the left to make sure the sizes work out. Both arrays
	// must be keyBytes long, and the output must be 2*keyBytes long.
	rBytesPadded := make([]byte, keyBytes)
	copy(rBytesPadded[keyBytes-len(r):], r)
	sBytesPadded := make([]byte, keyBytes)
	copy(sBytesPadded[keyBytes-len(s):], s)
	signatureEnc := append(rBytesPadded, sBytesPadded...)

	return base64.RawURLEncoding.EncodeToString(signatureEnc), nil
}

func retrieveCurveBits(alg string) (int, error) {
	var curveBits int
	switch alg {
	case "ECDSA_SHA_256":
		curveBits = 256
	case "ECDSA_SHA_384":
		curveBits = 384
	case "ECDSA_SHA_512":
		curveBits = 512
	default:
		return 0, fmt.Errorf("unsupported sign algorithm %s", alg)
	}
	return curveBits, nil
}

func messageDigest(message []byte, alg string) ([]byte, error) {
	var digest hash.Hash

	switch alg {
	case "ECDSA_SHA_256":
		digest = sha256.New()
	case "ECDSA_SHA_384":
		digest = sha512.New384()
	case "ECDSA_SHA_512":
		digest = sha512.New()
	default:
		return []byte{}, fmt.Errorf("unsupported sign algorithm %s", alg)
	}

	digest.Write(message)
	return digest.Sum(nil), nil
}

// oauth2ClientCredentialsAuthPlugin represents authentication via a bearer token in the HTTP Authorization header
// obtained through the OAuth2 client credentials flow
type oauth2ClientCredentialsAuthPlugin struct {
	GrantType            string                 `json:"grant_type"`
	TokenURL             string                 `json:"token_url"`
	ClientID             string                 `json:"client_id"`
	ClientSecret         string                 `json:"client_secret"`
	SigningKeyID         string                 `json:"signing_key"`
	Thumbprint           string                 `json:"thumbprint"`
	Claims               map[string]interface{} `json:"additional_claims"`
	IncludeJti           bool                   `json:"include_jti_claim"`
	Scopes               []string               `json:"scopes,omitempty"`
	AdditionalHeaders    map[string]string      `json:"additional_headers,omitempty"`
	AdditionalParameters map[string]string      `json:"additional_parameters,omitempty"`
	AWSKmsKey            *awsKmsKeyConfig       `json:"aws_kms,omitempty"`
	AWSSigningPlugin     *awsSigningAuthPlugin  `json:"aws_signing,omitempty"`
	ClientAssertionType  string                 `json:"client_assertion_type"`
	ClientAssertion      string                 `json:"client_assertion"`
	ClientAssertionPath  string                 `json:"client_assertion_path"`

	signingKey       *keys.Config
	signingKeyParsed interface{}
	tokenCache       *oauth2Token
	tlsSkipVerify    bool
	logger           logging.Logger
}

type oauth2Token struct {
	Token     string
	ExpiresAt time.Time
}

func (ap *oauth2ClientCredentialsAuthPlugin) createAuthJWT(ctx context.Context, extClaims map[string]interface{}, signingKey interface{}) (*string, error) {
	now := time.Now()
	claims := map[string]interface{}{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
	}
	for k, v := range extClaims {
		claims[k] = v
	}

	if len(ap.Scopes) > 0 {
		claims["scope"] = strings.Join(ap.Scopes, " ")
	}

	if ap.IncludeJti {
		jti, err := uuid.New(rand.Reader)
		if err != nil {
			return nil, err
		}
		claims["jti"] = jti
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var jwsHeaders []byte
	var signatureAlg string
	if ap.AWSKmsKey == nil {
		signatureAlg = ap.signingKey.Algorithm
	} else {
		signatureAlg, err = ap.mapKMSAlgToSign(ap.AWSKmsKey.Algorithm)
		if err != nil {
			return nil, err
		}
	}
	if ap.Thumbprint != "" {
		bytes, err := hex.DecodeString(ap.Thumbprint)
		if err != nil {
			return nil, err
		}
		x5t := base64.URLEncoding.EncodeToString(bytes)
		jwsHeaders = []byte(fmt.Sprintf(`{"typ":"JWT","alg":"%s","x5t":"%s"}`, signatureAlg, x5t))
	} else {
		jwsHeaders = []byte(fmt.Sprintf(`{"typ":"JWT","alg":"%s"}`, signatureAlg))
	}
	var jwsCompact []byte
	if ap.AWSKmsKey == nil {
		jwsCompact, err = jws.SignLiteral(payload,
			jwa.SignatureAlgorithm(signatureAlg),
			signingKey,
			jwsHeaders,
			rand.Reader)
	} else {
		jwsCompact, err = ap.SignWithKMS(ctx, payload, jwsHeaders)
	}
	if err != nil {
		return nil, err
	}
	jwt := string(jwsCompact)

	return &jwt, nil
}

func (*oauth2ClientCredentialsAuthPlugin) mapKMSAlgToSign(alg string) (string, error) {
	switch alg {
	case "ECDSA_SHA_256":
		return "ES256", nil
	case "ECDSA_SHA_384":
		return "ES384", nil
	case "ECDSA_SHA_512":
		return "ES512", nil
	default:
		return "", fmt.Errorf("unsupported sign algorithm %s", alg)
	}
}

// SignWithKMS will sign the JWT in AWS using the key stored in the supplied kmsArn
func (ap *oauth2ClientCredentialsAuthPlugin) SignWithKMS(ctx context.Context, payload []byte, hdrBuf []byte) ([]byte, error) {

	encodedHdr := base64.RawURLEncoding.EncodeToString(hdrBuf)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	input := encodedHdr + "." + encodedPayload
	digest, err := messageDigest([]byte(input), ap.AWSKmsKey.Algorithm)
	if err != nil {
		return nil, err
	}
	if ap.AWSSigningPlugin != nil {
		signature, err := ap.AWSSigningPlugin.SignDigest(ctx, digest, ap.AWSKmsKey.Name, ap.AWSKmsKey.Algorithm)
		if err != nil {
			return nil, err
		}
		der, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			return nil, err
		}
		signatureData, err := convertSignatureToBase64(ap.AWSKmsKey.Algorithm, der)
		if err != nil {
			return nil, err
		}

		signedAssertion := input + "." + signatureData

		return []byte(signedAssertion), nil
	}
	return nil, errors.New("missing AWS credentials, failed to sign the assertion with kms")
}

func (ap *oauth2ClientCredentialsAuthPlugin) parseSigningKey(c Config) (err error) {
	if ap.SigningKeyID == "" {
		return errors.New("signing_key required for jwt_bearer grant type")
	}

	if val, ok := c.keys[ap.SigningKeyID]; ok {
		if val.PrivateKey == "" {
			return errors.New("referenced signing_key does not include a private key")
		}
		ap.signingKey = val
	} else {
		return errors.New("signing_key refers to non-existent key")
	}

	alg := jwa.SignatureAlgorithm(ap.signingKey.Algorithm)
	ap.signingKeyParsed, err = sign.GetSigningKey(ap.signingKey.PrivateKey, alg)
	if err != nil {
		return err
	}

	return nil
}

func (ap *oauth2ClientCredentialsAuthPlugin) NewClient(c Config) (*http.Client, error) {
	t, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}

	if ap.GrantType == "" {
		// Use client_credentials as default to not break existing config
		ap.GrantType = grantTypeClientCredentials
	} else if ap.GrantType != grantTypeClientCredentials && ap.GrantType != grantTypeJwtBearer {
		return nil, errors.New("grant_type must be either client_credentials or jwt_bearer")
	}

	if ap.GrantType == grantTypeJwtBearer || (ap.GrantType == grantTypeClientCredentials && ap.SigningKeyID != "") {
		if err = ap.parseSigningKey(c); err != nil {
			return nil, err
		}
	}

	// Inherit skip verify from the "parent" settings. Should this be configurable on the credentials too?
	ap.tlsSkipVerify = c.AllowInsecureTLS

	ap.logger = c.logger

	if !strings.HasPrefix(ap.TokenURL, "https://") {
		return nil, errors.New("token_url required to use https scheme")
	}
	if ap.GrantType == grantTypeClientCredentials {
		clientCredentialExists := make(map[string]bool)
		clientCredentialExists["client_secret"] = ap.ClientSecret != ""
		clientCredentialExists["signing_key"] = ap.SigningKeyID != ""
		clientCredentialExists["aws_kms"] = ap.AWSKmsKey != nil
		clientCredentialExists["client_assertion"] = ap.ClientAssertion != ""
		clientCredentialExists["client_assertion_path"] = ap.ClientAssertionPath != ""

		var notEmptyVarCount int

		for _, credentialSet := range clientCredentialExists {
			if credentialSet {
				notEmptyVarCount++
			}
		}

		if notEmptyVarCount == 0 {
			return nil, errors.New("please provide one of client_secret, signing_key, aws_kms, client_assertion, or client_assertion_path required")
		}

		if notEmptyVarCount > 1 {
			return nil, errors.New("can only use one of client_secret, signing_key, aws_kms, client_assertion, or client_assertion_path")
		}

		if clientCredentialExists["aws_kms"] {
			if ap.AWSSigningPlugin == nil {
				return nil, errors.New("aws_kms and aws_signing required")
			}
			// initialize the awsSigningAuthPlugin
			_, err = ap.AWSSigningPlugin.NewClient(c)
			if err != nil {
				return nil, err
			}
		} else if clientCredentialExists["client_assertion"] {
			if ap.ClientAssertionType == "" {
				ap.ClientAssertionType = defaultClientAssertionType
			}
			if ap.ClientID == "" {
				return nil, errors.New("client_id and client_assertion required")
			}
		} else if clientCredentialExists["client_assertion_path"] {
			if ap.ClientAssertionType == "" {
				ap.ClientAssertionType = defaultClientAssertionType
			}
			if ap.ClientID == "" {
				return nil, errors.New("client_id and client_assertion_path required")
			}
		} else if clientCredentialExists["client_secret"] {
			if ap.ClientID == "" {
				return nil, errors.New("client_id and client_secret required")
			}
		}
	}

	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

// requestToken tries to obtain an access token using either the client credentials flow
// https://tools.ietf.org/html/rfc6749#section-4.4
// or the JWT authorization grant
// https://tools.ietf.org/html/rfc7523
func (ap *oauth2ClientCredentialsAuthPlugin) requestToken(ctx context.Context) (*oauth2Token, error) {
	body := url.Values{}
	if ap.GrantType == grantTypeJwtBearer {
		authJwt, err := ap.createAuthJWT(ctx, ap.Claims, ap.signingKeyParsed)
		if err != nil {
			return nil, err
		}
		body.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
		body.Add("assertion", *authJwt)
	} else {
		body.Add("grant_type", grantTypeClientCredentials)

		if ap.SigningKeyID != "" || ap.AWSKmsKey != nil {
			authJwt, err := ap.createAuthJWT(ctx, ap.Claims, ap.signingKeyParsed)
			if err != nil {
				return nil, err
			}
			body.Add("client_assertion_type", defaultClientAssertionType)
			body.Add("client_assertion", *authJwt)

			if ap.ClientID != "" {
				body.Add("client_id", ap.ClientID)
			}
		} else if ap.ClientAssertion != "" {
			if ap.ClientAssertionType == "" {
				ap.ClientAssertionType = defaultClientAssertionType
			}
			if ap.ClientID != "" {
				body.Add("client_id", ap.ClientID)
			}
			body.Add("client_assertion_type", ap.ClientAssertionType)
			body.Add("client_assertion", ap.ClientAssertion)
		} else if ap.ClientAssertionPath != "" {
			if ap.ClientAssertionType == "" {
				ap.ClientAssertionType = defaultClientAssertionType
			}
			bytes, err := os.ReadFile(ap.ClientAssertionPath)
			if err != nil {
				return nil, err
			}
			if ap.ClientID != "" {
				body.Add("client_id", ap.ClientID)
			}
			body.Add("client_assertion_type", ap.ClientAssertionType)
			body.Add("client_assertion", strings.TrimSpace(string(bytes)))
		}
	}

	if len(ap.Scopes) > 0 {
		body.Add("scope", strings.Join(ap.Scopes, " "))
	}

	for k, v := range ap.AdditionalParameters {
		body.Set(k, v)
	}

	r, err := http.NewRequestWithContext(ctx, "POST", ap.TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if ap.GrantType == grantTypeClientCredentials && ap.ClientSecret != "" {
		r.SetBasicAuth(ap.ClientID, ap.ClientSecret)
	}

	for k, v := range ap.AdditionalHeaders {
		r.Header.Add(k, v)
	}

	client := DefaultRoundTripperClient(&tls.Config{InsecureSkipVerify: ap.tlsSkipVerify}, 10)
	response, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyRaw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error in response from OAuth2 token endpoint: %v", string(bodyRaw))
	}

	var tokenResponse tokenEndpointResponse
	err = json.Unmarshal(bodyRaw, &tokenResponse)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(tokenResponse.TokenType, "bearer") {
		return nil, errors.New("unknown token type returned from token endpoint")
	}

	return &oauth2Token{
		Token:     strings.TrimSpace(tokenResponse.AccessToken),
		ExpiresAt: time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}, nil
}

func (ap *oauth2ClientCredentialsAuthPlugin) Prepare(req *http.Request) error {
	minTokenLifetime := float64(10)
	if ap.tokenCache == nil || time.Until(ap.tokenCache.ExpiresAt).Seconds() < minTokenLifetime {
		ap.logger.Debug("Requesting token from token_url %v", ap.TokenURL)
		token, err := ap.requestToken(req.Context())
		if err != nil {
			return err
		}
		ap.tokenCache = token
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", ap.tokenCache.Token))
	return nil
}

// clientTLSAuthPlugin represents authentication via client certificate on a TLS connection
type clientTLSAuthPlugin struct {
	Cert                 string `json:"cert"`
	PrivateKey           string `json:"private_key"`
	PrivateKeyPassphrase string `json:"private_key_passphrase,omitempty"`
	CACert               string `json:"ca_cert,omitempty"`            // Deprecated: Use `services[_].tls.ca_cert` instead
	SystemCARequired     bool   `json:"system_ca_required,omitempty"` // Deprecated: Use `services[_].tls.system_ca_required` instead
}

func (ap *clientTLSAuthPlugin) NewClient(c Config) (*http.Client, error) {
	tlsConfig, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}

	if ap.Cert == "" {
		return nil, errors.New("client certificate is needed when client TLS is enabled")
	}
	if ap.PrivateKey == "" {
		return nil, errors.New("private key is needed when client TLS is enabled")
	}

	var keyPEMBlock []byte
	data, err := os.ReadFile(ap.PrivateKey)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM data could not be found")
	}

	// nolint: staticcheck // We don't want to forbid users from using this encryption.
	if x509.IsEncryptedPEMBlock(block) {
		if ap.PrivateKeyPassphrase == "" {
			return nil, errors.New("client certificate passphrase is needed, because the certificate is password encrypted")
		}
		// nolint: staticcheck // We don't want to forbid users from using this encryption.
		block, err := x509.DecryptPEMBlock(block, []byte(ap.PrivateKeyPassphrase))
		if err != nil {
			return nil, err
		}
		key, err := x509.ParsePKCS8PrivateKey(block)
		if err != nil {
			key, err = x509.ParsePKCS1PrivateKey(block)
			if err != nil {
				return nil, fmt.Errorf("private key should be a PEM or plain PKCS1 or PKCS8; parse error: %v", err)
			}
		}
		rsa, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is invalid")
		}
		keyPEMBlock = pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(rsa),
			},
		)
	} else {
		keyPEMBlock = data
	}

	certPEMBlock, err := os.ReadFile(ap.Cert)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	var client *http.Client

	if c.TLS != nil && c.TLS.CACert != "" {
		client = DefaultRoundTripperClient(tlsConfig, *c.ResponseHeaderTimeoutSeconds)
	} else {
		if ap.CACert != "" {
			c.logger.Warn("Deprecated 'services[_].credentials.client_tls.ca_cert' configuration specified. Use 'services[_].tls.ca_cert' instead. See https://www.openpolicyagent.org/docs/latest/configuration/#services")
			caCert, err := os.ReadFile(ap.CACert)
			if err != nil {
				return nil, err
			}

			var caCertPool *x509.CertPool
			if ap.SystemCARequired {
				caCertPool, err = x509.SystemCertPool()
				if err != nil {
					return nil, err
				}
			} else {
				caCertPool = x509.NewCertPool()
			}

			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, errors.New("unable to parse and append CA certificate to certificate pool")
			}
			tlsConfig.RootCAs = caCertPool
		}

		client = DefaultRoundTripperClient(tlsConfig, *c.ResponseHeaderTimeoutSeconds)
	}

	return client, nil
}

func (*clientTLSAuthPlugin) Prepare(_ *http.Request) error {
	return nil
}

// awsSigningAuthPlugin represents authentication using AWS V4 HMAC signing in the Authorization header
type awsSigningAuthPlugin struct {
	AWSEnvironmentCredentials *awsEnvironmentCredentialService `json:"environment_credentials,omitempty"`
	AWSMetadataCredentials    *awsMetadataCredentialService    `json:"metadata_credentials,omitempty"`
	AWSAssumeRoleCredentials  *awsAssumeRoleCredentialService  `json:"assume_role_credentials,omitempty"`
	AWSWebIdentityCredentials *awsWebIdentityCredentialService `json:"web_identity_credentials,omitempty"`
	AWSProfileCredentials     *awsProfileCredentialService     `json:"profile_credentials,omitempty"`

	AWSService          string `json:"service,omitempty"`
	AWSSignatureVersion string `json:"signature_version,omitempty"`

	host          string
	ecrAuthPlugin *ecrAuthPlugin
	kmsSignPlugin *awsKMSSignPlugin

	logger logging.Logger
}

type awsCredentialServiceChain struct {
	awsCredentialServices []awsCredentialService
	logger                logging.Logger
}

func (acs *awsCredentialServiceChain) addService(service awsCredentialService) {
	acs.awsCredentialServices = append(acs.awsCredentialServices, service)
}

type awsCredentialCheckErrors []*awsCredentialCheckError

func (e awsCredentialCheckErrors) Error() string {

	if len(e) == 0 {
		return "no error(s)"
	}

	if len(e) == 1 {
		return fmt.Sprintf("1 error occurred: %v", e[0].Error())
	}

	s := make([]string, len(e))
	for i, err := range e {
		s[i] = err.Error()
	}

	return fmt.Sprintf("%d errors occurred:\n%s", len(e), strings.Join(s, "\n"))
}

type awsCredentialCheckError struct {
	message string
}

func newAWSCredentialError(message string) *awsCredentialCheckError {
	return &awsCredentialCheckError{
		message: message,
	}
}

func (e *awsCredentialCheckError) Error() string {
	return e.message
}

func (acs *awsCredentialServiceChain) credentials(ctx context.Context) (aws.Credentials, error) {
	var errs awsCredentialCheckErrors

	for _, service := range acs.awsCredentialServices {
		credential, err := service.credentials(ctx)
		if err != nil {
			acs.logger.Debug("awsSigningAuthPlugin:%T failed: %v", service, err)

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return aws.Credentials{}, err
			}

			errs = append(errs, newAWSCredentialError(err.Error()))
			continue
		}

		acs.logger.Debug("awsSigningAuthPlugin:%T successful", service)
		return credential, nil
	}

	return aws.Credentials{}, fmt.Errorf("all AWS credential providers failed: %v", errs)
}

func (ap *awsSigningAuthPlugin) awsCredentialService() awsCredentialService {
	chain := awsCredentialServiceChain{
		logger: ap.logger,
	}

	/*
		Here we maintain the order of addition to the chain inline with
		the order of credential providers followed by default by the
		AWS SDK. For example

		https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/auth/DefaultAWSCredentialsProviderChain.html
	*/

	if ap.AWSEnvironmentCredentials != nil {
		ap.AWSEnvironmentCredentials.logger = ap.logger
		chain.addService(ap.AWSEnvironmentCredentials)
	}

	if ap.AWSAssumeRoleCredentials != nil {
		ap.AWSAssumeRoleCredentials.logger = ap.logger
		chain.addService(ap.AWSAssumeRoleCredentials)
	}

	if ap.AWSWebIdentityCredentials != nil {
		ap.AWSWebIdentityCredentials.logger = ap.logger
		chain.addService(ap.AWSWebIdentityCredentials)
	}

	if ap.AWSProfileCredentials != nil {
		ap.AWSProfileCredentials.logger = ap.logger
		chain.addService(ap.AWSProfileCredentials)
	}

	if ap.AWSMetadataCredentials != nil {
		ap.AWSMetadataCredentials.logger = ap.logger
		chain.addService(ap.AWSMetadataCredentials)
	}

	return &chain
}

func (ap *awsSigningAuthPlugin) NewClient(c Config) (*http.Client, error) {
	t, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}

	ap.host = url.Host

	if ap.logger == nil {
		ap.logger = c.logger
	}

	if err := ap.validateAndSetDefaults(c.Type); err != nil {
		return nil, err
	}

	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

func (ap *awsSigningAuthPlugin) Prepare(req *http.Request) error {
	if ap.host != req.URL.Host {
		// Return early if the host does not match.
		// This can happen when the OCI registry responded with a redirect to another host.
		// For instance, ECR redirects to S3 and the ECR auth header should not be included in the S3 request.
		return nil
	}

	switch ap.AWSService {
	case "ecr":
		return ap.ecrAuthPlugin.Prepare(req)
	default:
		creds, err := ap.awsCredentialService().credentials(req.Context())
		if err != nil {
			return fmt.Errorf("failed to get aws credentials: %w", err)
		}

		ap.logger.Debug("Signing request with AWS credentials.")

		return aws.SignRequest(req, ap.AWSService, creds, time.Now(), ap.AWSSignatureVersion)
	}
}

func (ap *awsSigningAuthPlugin) validateAndSetDefaults(serviceType string) error {
	cfgs := map[bool]int{}
	cfgs[ap.AWSEnvironmentCredentials != nil]++
	cfgs[ap.AWSMetadataCredentials != nil]++
	cfgs[ap.AWSAssumeRoleCredentials != nil]++
	cfgs[ap.AWSWebIdentityCredentials != nil]++
	cfgs[ap.AWSProfileCredentials != nil]++

	if cfgs[true] == 0 {
		return errors.New("a AWS credential service must be specified when S3 signing is enabled")
	}

	if ap.AWSMetadataCredentials != nil {
		if ap.AWSMetadataCredentials.RegionName == "" {
			return errors.New("at least aws_region must be specified for AWS metadata credential service")
		}
	}

	if ap.AWSAssumeRoleCredentials != nil {
		if err := ap.AWSAssumeRoleCredentials.populateFromEnv(); err != nil {
			return err
		}
	}

	if ap.AWSWebIdentityCredentials != nil {
		if err := ap.AWSWebIdentityCredentials.populateFromEnv(); err != nil {
			return err
		}
	}

	ap.AWSService = strings.ToLower(ap.AWSService)

	// Only allow ECR for OCI service types
	if serviceType == "oci" {
		if ap.AWSService == "" {
			ap.AWSService = "ecr"
		}

		if ap.AWSService != "ecr" {
			return fmt.Errorf(`cannot use aws service %q with service type "oci"`, ap.AWSService)
		}

		// We need to setup a special auth plugin for ECR.
		ap.ecrAuthPlugin = newECRAuthPlugin(ap)
	} else {
		// Disallow ECR for non-OCI service types
		if ap.AWSService == "ecr" {
			return errors.New(`aws service "ecr" must be used with service type "oci"`)
		}
		if ap.AWSService == "kms" && ap.kmsSignPlugin == nil {
			// We need a special plugin for KMS.
			ap.kmsSignPlugin = newKMSSignPlugin(ap)
		}
		if ap.AWSService == "" {
			ap.AWSService = awsSigv4SigningDefaultService
		}
	}

	if ap.AWSSignatureVersion == "" {
		ap.AWSSignatureVersion = "4"
	}

	return nil
}

func (ap *awsSigningAuthPlugin) SignDigest(ctx context.Context, digest []byte, keyID string, signingAlgorithm string) (string, error) {
	switch ap.AWSService {
	case "kms":
		return ap.kmsSignPlugin.SignDigest(ctx, digest, keyID, signingAlgorithm)
	default:
		return "", fmt.Errorf(`cannot use SignDigest with aws service %q`, ap.AWSService)
	}
}
