/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package token

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/square/go-jose.v2"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

const (
	AccessToken       Type   = "access_token"
	RefreshToken      Type   = "refresh_token"
	StaticToken       Type   = "static_token"
	AuthorizationCode Type   = "code"
	IDToken           Type   = "id_token"
	headerKeyID       string = "kid"
	headerAlgorithm   string = "alg"
)

type Type string

type IssueRequest struct {
	User      user.Info
	ExpiresIn time.Duration
	Claims
}

type VerifiedResponse struct {
	User user.Info
	Claims
}

// Keys hold encryption and signing keys.
type Keys struct {
	SigningKey    *jose.JSONWebKey
	SigningKeyPub *jose.JSONWebKey
}

// Issuer issues token to user, tokens are required to perform mutating requests to resources
type Issuer interface {
	// IssueTo issues a token a User, return error if issuing process failed
	IssueTo(request *IssueRequest) (string, error)

	// Verify verifies a token, and return a user info if it's a valid token, otherwise return error
	Verify(string) (*VerifiedResponse, error)

	// Keys hold encryption and signing keys.
	Keys() *Keys
}

type Claims struct {
	jwt.RegisteredClaims
	// Private Claim Names
	// TokenType defined the type of the token
	TokenType Type `json:"token_type,omitempty"`
	// Username user identity, deprecated field
	Username string `json:"username,omitempty"`
	// Extra contains the additional information
	Extra map[string][]string `json:"extra,omitempty"`

	// Used for issuing authorization code
	// Scopes can be used to request that specific sets of information be made available as Claim Values.
	Scopes []string `json:"scopes,omitempty"`

	// The following is well-known ID Token fields

	// End-User's full url in displayable form including all url parts,
	// possibly including titles and suffixes, ordered according to the End-User's locale and preferences.
	Name string `json:"url,omitempty"`
	// String value used to associate a Client session with an ID Token, and to mitigate replay attacks.
	// The value is passed through unmodified from the Authentication Request to the ID Token.
	Nonce string `json:"nonce,omitempty"`
	// End-User's preferred e-mail address.
	Email string `json:"email,omitempty"`
	// End-User's locale, represented as a BCP47 [RFC5646] language tag.
	Locale string `json:"locale,omitempty"`
	// Shorthand url by which the End-User wishes to be referred to at the RP,
	PreferredUsername string `json:"preferred_username,omitempty"`
}

type issuer struct {
	// Issuer Identifier
	url string
	// signing access_token and refresh_token
	secret []byte
	// signing id_token
	signKey *Keys
	// Token verification maximum time difference
	maximumClockSkew time.Duration
}

func (s *issuer) IssueTo(request *IssueRequest) (string, error) {
	issueAt := time.Now()
	claims := Claims{
		Username:  request.User.GetName(),
		Extra:     request.User.GetExtra(),
		TokenType: request.TokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issueAt),
			Subject:  request.User.GetName(),
			Issuer:   s.url,
		},
	}

	if len(request.Audience) > 0 {
		claims.Audience = request.Audience
	}
	if request.Name != "" {
		claims.Name = request.Name
	}
	if request.Nonce != "" {
		claims.Nonce = request.Nonce
	}
	if request.Email != "" {
		claims.Email = request.Email
	}
	if request.PreferredUsername != "" {
		claims.PreferredUsername = request.PreferredUsername
	}
	if request.Locale != "" {
		claims.Locale = request.Locale
	}
	if len(request.Scopes) > 0 {
		claims.Scopes = request.Scopes
	}
	if request.ExpiresIn > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(issueAt.Add(request.ExpiresIn))
	}

	var token string
	var err error
	if request.TokenType == IDToken {
		t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		t.Header[headerKeyID] = s.signKey.SigningKey.KeyID
		token, err = t.SignedString(s.signKey.SigningKey.Key)
	} else {
		token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	}
	if err != nil {
		klog.Warningf("jwt: failed to issue token: %v", err)
		return "", err
	}
	return token, nil
}

func (s *issuer) Verify(token string) (*VerifiedResponse, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg(), jwt.SigningMethodRS256.Alg()}),
		jwt.WithoutClaimsValidation())

	var claims Claims
	_, err := parser.ParseWithClaims(token, &claims, s.keyFunc)
	if err != nil {
		klog.Warningf("jwt: failed to parse token: %v", err)
		return nil, err
	}

	now := time.Now()
	if !claims.VerifyExpiresAt(now, false) {
		delta := now.Sub(claims.ExpiresAt.Time)
		err = fmt.Errorf("jwt: token is expired by %v", delta)
		klog.V(4).Info(err)
		return nil, err
	}

	// allowing a clock skew when checking the time-based values.
	skewedTime := now.Add(s.maximumClockSkew)
	if !claims.VerifyIssuedAt(skewedTime, false) {
		err = fmt.Errorf("jwt: token used before issued, iat:%v, now:%v", claims.IssuedAt, now)
		klog.Warning(err)
		return nil, err
	}

	verified := &VerifiedResponse{
		User: &user.DefaultInfo{
			Name:  claims.Username,
			Extra: claims.Extra,
		},
		Claims: claims,
	}

	return verified, nil
}

func (s *issuer) Keys() *Keys {
	return s.signKey
}

func (s *issuer) keyFunc(token *jwt.Token) (i interface{}, err error) {
	alg, _ := token.Header[headerAlgorithm].(string)
	switch alg {
	case jwt.SigningMethodHS256.Alg():
		return s.secret, nil
	case jwt.SigningMethodRS256.Alg():
		return s.signKey.SigningKey.Key, nil
	default:
		return nil, fmt.Errorf("unexpect signature algorithm %v", token.Header[headerAlgorithm])
	}
}

func loadPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("private key not in pem format")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to key file: %v", err)
	}
	return key, nil
}

func generatePrivateKeyData() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	data := x509.MarshalPKCS1PrivateKey(privateKey)
	pemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: data,
		},
	)
	return pemData, nil
}

func loadSignKey(config *oauth.IssuerOptions) (*rsa.PrivateKey, string, error) {
	var signKey *rsa.PrivateKey
	var signKeyData []byte
	var err error

	if config.SignKey != "" {
		signKeyData, err = os.ReadFile(config.SignKey)
		if err != nil {
			klog.Errorf("issuer: failed to read private key file %s: %v", config.SignKey, err)
			return nil, "", err
		}
	} else if config.SignKeyData != "" {
		signKeyData, err = base64.StdEncoding.DecodeString(config.SignKeyData)
		if err != nil {
			klog.Errorf("issuer: failed to decode sign key data: %s", err)
			return nil, "", err
		}
	}

	// automatically generate private key
	if len(signKeyData) == 0 {
		signKeyData, err = generatePrivateKeyData()
		if err != nil {
			klog.Errorf("issuer: failed to generate private key: %v", err)
			return nil, "", err
		}
	}

	if len(signKeyData) > 0 {
		signKey, err = loadPrivateKey(signKeyData)
		if err != nil {
			klog.Errorf("issuer: failed to load private key from data: %v", err)
		}
	}

	keyID := fmt.Sprint(fnv32a(signKeyData))
	return signKey, keyID, nil
}

func NewIssuer(config *oauth.IssuerOptions) (Issuer, error) {
	// TODO(hongming) automatically rotates keys
	signKey, keyID, err := loadSignKey(config)
	if err != nil {
		return nil, err
	}
	return &issuer{
		url:              config.URL,
		secret:           []byte(config.JWTSecret),
		maximumClockSkew: config.MaximumClockSkew,
		signKey: &Keys{
			SigningKey: &jose.JSONWebKey{
				Key:       signKey,
				KeyID:     keyID,
				Algorithm: jwt.SigningMethodRS256.Alg(),
				Use:       "sig",
			},
			SigningKeyPub: &jose.JSONWebKey{
				Key:       signKey.Public(),
				KeyID:     keyID,
				Algorithm: jwt.SigningMethodRS256.Alg(),
				Use:       "sig",
			},
		},
	}, nil
}

// fnv32a hashes using fnv32a algorithm
func fnv32a(data []byte) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write(data)
	return algorithm.Sum32()
}
