/*
Copyright 2014 The Kubernetes Authors.

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

package serviceaccount

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// ServiceAccountTokenGetter defines functions to retrieve a named service account and secret
type ServiceAccountTokenGetter interface {
	GetServiceAccount(namespace, name string) (*v1.ServiceAccount, error)
	GetPod(namespace, name string) (*v1.Pod, error)
	GetSecret(namespace, name string) (*v1.Secret, error)
}

type TokenGenerator interface {
	// GenerateToken generates a token which will identify the given
	// ServiceAccount. privateClaims is an interface that will be
	// serialized into the JWT payload JSON encoding at the root level of
	// the payload object. Public claims take precedent over private
	// claims i.e. if both claims and privateClaims have an "exp" field,
	// the value in claims will be used.
	GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error)
}

// JWTTokenGenerator returns a TokenGenerator that generates signed JWT tokens, using the given privateKey.
// privateKey is a PEM-encoded byte array of a private RSA key.
// JWTTokenAuthenticator()
func JWTTokenGenerator(iss string, privateKey interface{}) TokenGenerator {
	return &jwtTokenGenerator{
		iss:        iss,
		privateKey: privateKey,
	}
}

type jwtTokenGenerator struct {
	iss        string
	privateKey interface{}
}

func (j *jwtTokenGenerator) GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error) {
	var alg jose.SignatureAlgorithm
	switch privateKey := j.privateKey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		switch privateKey.Curve {
		case elliptic.P256():
			alg = jose.ES256
		case elliptic.P384():
			alg = jose.ES384
		case elliptic.P521():
			alg = jose.ES512
		default:
			return "", fmt.Errorf("unknown private key curve, must be 256, 384, or 521")
		}
	default:
		return "", fmt.Errorf("unknown private key type %T, must be *rsa.PrivateKey or *ecdsa.PrivateKey", j.privateKey)
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key:       j.privateKey,
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	// claims are applied in reverse precedence
	return jwt.Signed(signer).
		Claims(privateClaims).
		Claims(claims).
		Claims(&jwt.Claims{
			Issuer: j.iss,
		}).
		CompactSerialize()
}

// JWTTokenAuthenticator authenticates tokens as JWT tokens produced by JWTTokenGenerator
// Token signatures are verified using each of the given public keys until one works (allowing key rotation)
// If lookup is true, the service account and secret referenced as claims inside the token are retrieved and verified with the provided ServiceAccountTokenGetter
func JWTTokenAuthenticator(iss string, keys []interface{}, validator Validator) authenticator.Token {
	return &jwtTokenAuthenticator{
		iss:       iss,
		keys:      keys,
		validator: validator,
	}
}

type jwtTokenAuthenticator struct {
	iss       string
	keys      []interface{}
	validator Validator
}

// Validator is called by the JWT token authentictaor to apply domain specific
// validation to a token and extract user information.
type Validator interface {
	// Validate validates a token and returns user information or an error.
	// Validator can assume that the issuer and signature of a token are already
	// verified when this function is called.
	Validate(tokenData string, public *jwt.Claims, private interface{}) (namespace, name, uid string, err error)
	// NewPrivateClaims returns a struct that the authenticator should
	// deserialize the JWT payload into. The authenticator may then pass this
	// struct back to the Validator as the 'private' argument to a Validate()
	// call. This struct should contain fields for any private claims that the
	// Validator requires to validate the JWT.
	NewPrivateClaims() interface{}
}

func (j *jwtTokenAuthenticator) AuthenticateToken(tokenData string) (user.Info, bool, error) {
	if !j.hasCorrectIssuer(tokenData) {
		return nil, false, nil
	}

	tok, err := jwt.ParseSigned(tokenData)
	if err != nil {
		return nil, false, nil
	}

	public := &jwt.Claims{}
	private := j.validator.NewPrivateClaims()

	var (
		found   bool
		errlist []error
	)
	for _, key := range j.keys {
		if err := tok.Claims(key, public, private); err != nil {
			errlist = append(errlist, err)
			continue
		}
		found = true
		break
	}

	if !found {
		return nil, false, utilerrors.NewAggregate(errlist)
	}

	// If we get here, we have a token with a recognized signature and
	// issuer string.
	ns, name, uid, err := j.validator.Validate(tokenData, public, private)
	if err != nil {
		return nil, false, err
	}

	return UserInfo(ns, name, uid), true, nil
}

// hasCorrectIssuer returns true if tokenData is a valid JWT in compact
// serialization format and the "iss" claim matches the iss field of this token
// authenticator, and otherwise returns false.
//
// Note: go-jose currently does not allow access to unverified JWS payloads.
// See https://github.com/square/go-jose/issues/169
func (j *jwtTokenAuthenticator) hasCorrectIssuer(tokenData string) bool {
	parts := strings.Split(tokenData, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	claims := struct {
		// WARNING: this JWT is not verified. Do not trust these claims.
		Issuer string `json:"iss"`
	}{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	if claims.Issuer != j.iss {
		return false
	}
	return true

}
