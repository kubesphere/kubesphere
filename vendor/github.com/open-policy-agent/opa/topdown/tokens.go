// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/jwx/jwk"
	"github.com/open-policy-agent/opa/internal/jwx/jws"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var (
	jwtEncKey = ast.StringTerm("enc")
	jwtCtyKey = ast.StringTerm("cty")
	jwtAlgKey = ast.StringTerm("alg")
	jwtIssKey = ast.StringTerm("iss")
	jwtExpKey = ast.StringTerm("exp")
	jwtNbfKey = ast.StringTerm("nbf")
	jwtAudKey = ast.StringTerm("aud")
)

// JSONWebToken represent the 3 parts (header, payload & signature) of
//              a JWT in Base64.
type JSONWebToken struct {
	header        string
	payload       string
	signature     string
	decodedHeader ast.Object
}

// decodeHeader populates the decodedHeader field.
func (token *JSONWebToken) decodeHeader() (err error) {
	var h ast.Value
	if h, err = builtinBase64UrlDecode(ast.String(token.header)); err != nil {
		return fmt.Errorf("JWT header had invalid encoding: %v", err)
	}
	if token.decodedHeader, err = validateJWTHeader(string(h.(ast.String))); err != nil {
		return err
	}
	return
}

// Implements JWT decoding/validation based on RFC 7519 Section 7.2:
// https://tools.ietf.org/html/rfc7519#section-7.2
// It does no data validation, it merely checks that the given string
// represents a structurally valid JWT. It supports JWTs using JWS compact
// serialization.
func builtinJWTDecode(a ast.Value) (ast.Value, error) {
	token, err := decodeJWT(a)
	if err != nil {
		return nil, err
	}

	if err = token.decodeHeader(); err != nil {
		return nil, err
	}

	p, err := builtinBase64UrlDecode(ast.String(token.payload))
	if err != nil {
		return nil, fmt.Errorf("JWT payload had invalid encoding: %v", err)
	}

	if cty := token.decodedHeader.Get(jwtCtyKey); cty != nil {
		ctyVal := string(cty.Value.(ast.String))
		// It is possible for the contents of a token to be another
		// token as a result of nested signing or encryption. To handle
		// the case where we are given a token such as this, we check
		// the content type and recurse on the payload if the content
		// is "JWT".
		// When the payload is itself another encoded JWT, then its
		// contents are quoted (behavior of https://jwt.io/). To fix
		// this, remove leading and trailing quotes.
		if ctyVal == "JWT" {
			p, err = builtinTrim(p, ast.String(`"'`))
			if err != nil {
				panic("not reached")
			}
			return builtinJWTDecode(p)
		}
	}

	payload, err := extractJSONObject(string(p.(ast.String)))
	if err != nil {
		return nil, err
	}

	s, err := builtinBase64UrlDecode(ast.String(token.signature))
	if err != nil {
		return nil, fmt.Errorf("JWT signature had invalid encoding: %v", err)
	}
	sign := hex.EncodeToString([]byte(s.(ast.String)))

	arr := []*ast.Term{
		ast.NewTerm(token.decodedHeader),
		ast.NewTerm(payload),
		ast.StringTerm(sign),
	}

	return ast.NewArray(arr...), nil
}

// Implements RS256 JWT signature verification
func builtinJWTVerifyRS256(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha256.New, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPKCS1v15(
			publicKey,
			crypto.SHA256,
			digest,
			signature)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements RS384 JWT signature verification
func builtinJWTVerifyRS384(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha512.New384, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPKCS1v15(
			publicKey,
			crypto.SHA384,
			digest,
			signature)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements RS512 JWT signature verification
func builtinJWTVerifyRS512(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha512.New, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPKCS1v15(
			publicKey,
			crypto.SHA512,
			digest,
			signature)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements PS256 JWT signature verification
func builtinJWTVerifyPS256(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha256.New, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPSS(
			publicKey,
			crypto.SHA256,
			digest,
			signature,
			nil)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements PS384 JWT signature verification
func builtinJWTVerifyPS384(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha512.New384, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPSS(
			publicKey,
			crypto.SHA384,
			digest,
			signature,
			nil)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements PS512 JWT signature verification
func builtinJWTVerifyPS512(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerifyRSA(args[0].Value, args[1].Value, sha512.New, func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error {
		return rsa.VerifyPSS(
			publicKey,
			crypto.SHA512,
			digest,
			signature,
			nil)
	})
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements RSA JWT signature verification.
func builtinJWTVerifyRSA(a ast.Value, b ast.Value, hasher func() hash.Hash, verify func(publicKey *rsa.PublicKey, digest []byte, signature []byte) error) (ast.Value, error) {
	return builtinJWTVerify(a, b, hasher, func(publicKey interface{}, digest []byte, signature []byte) error {
		publicKeyRsa, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("incorrect public key type")
		}
		return verify(publicKeyRsa, digest, signature)
	})
}

// Implements ES256 JWT signature verification.
func builtinJWTVerifyES256(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerify(args[0].Value, args[1].Value, sha256.New, verifyES)
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements ES384 JWT signature verification
func builtinJWTVerifyES384(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerify(args[0].Value, args[1].Value, sha512.New384, verifyES)
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

// Implements ES512 JWT signature verification
func builtinJWTVerifyES512(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	result, err := builtinJWTVerify(args[0].Value, args[1].Value, sha512.New, verifyES)
	if err == nil {
		return iter(ast.NewTerm(result))
	}
	return err
}

func verifyES(publicKey interface{}, digest []byte, signature []byte) error {
	publicKeyEcdsa, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("incorrect public key type")
	}
	r, s := &big.Int{}, &big.Int{}
	n := len(signature) / 2
	r.SetBytes(signature[:n])
	s.SetBytes(signature[n:])
	if ecdsa.Verify(publicKeyEcdsa, digest, r, s) {
		return nil
	}
	return fmt.Errorf("ECDSA signature verification error")
}

// getKeyFromCertOrJWK returns the public key found in a X.509 certificate or JWK key(s).
// A valid PEM block is never valid JSON (and vice versa), hence can try parsing both.
func getKeyFromCertOrJWK(certificate string) ([]interface{}, error) {
	if block, rest := pem.Decode([]byte(certificate)); block != nil {
		if len(rest) > 0 {
			return nil, fmt.Errorf("extra data after a PEM certificate block")
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse a PEM certificate")
			}

			return []interface{}{cert.PublicKey}, nil
		}

		if block.Type == "PUBLIC KEY" {
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse a PEM public key")
			}

			return []interface{}{key}, nil
		}

		return nil, fmt.Errorf("failed to extract a Key from the PEM certificate")
	}

	jwks, err := jwk.ParseString(certificate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse a JWK key (set)")
	}

	var keys []interface{}
	for _, k := range jwks.Keys {
		key, err := k.Materialize()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Implements JWT signature verification.
func builtinJWTVerify(a ast.Value, b ast.Value, hasher func() hash.Hash, verify func(publicKey interface{}, digest []byte, signature []byte) error) (ast.Value, error) {
	token, err := decodeJWT(a)
	if err != nil {
		return nil, err
	}

	s, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}

	keys, err := getKeyFromCertOrJWK(string(s))
	if err != nil {
		return nil, err
	}

	signature, err := token.decodeSignature()
	if err != nil {
		return nil, err
	}

	// Validate the JWT signature
	for _, key := range keys {
		err = verify(key,
			getInputSHA([]byte(token.header+"."+token.payload), hasher),
			[]byte(signature))

		if err == nil {
			return ast.Boolean(true), nil
		}
	}

	// None of the keys worked, return false
	return ast.Boolean(false), nil
}

// Implements HS256 (secret) JWT signature verification
func builtinJWTVerifyHS256(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	// Decode the JSON Web Token
	token, err := decodeJWT(args[0].Value)
	if err != nil {
		return err
	}

	// Process Secret input
	astSecret, err := builtins.StringOperand(args[1].Value, 2)
	if err != nil {
		return err
	}
	secret := string(astSecret)

	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write([]byte(token.header + "." + token.payload))
	if err != nil {
		return err
	}

	signature, err := token.decodeSignature()
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(ast.Boolean(hmac.Equal([]byte(signature), mac.Sum(nil)))))
}

// Implements HS384 JWT signature verification
func builtinJWTVerifyHS384(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	// Decode the JSON Web Token
	token, err := decodeJWT(args[0].Value)
	if err != nil {
		return err
	}

	// Process Secret input
	astSecret, err := builtins.StringOperand(args[1].Value, 2)
	if err != nil {
		return err
	}
	secret := string(astSecret)

	mac := hmac.New(sha512.New384, []byte(secret))
	_, err = mac.Write([]byte(token.header + "." + token.payload))
	if err != nil {
		return err
	}

	signature, err := token.decodeSignature()
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(ast.Boolean(hmac.Equal([]byte(signature), mac.Sum(nil)))))
}

// Implements HS512 JWT signature verification
func builtinJWTVerifyHS512(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	// Decode the JSON Web Token
	token, err := decodeJWT(args[0].Value)
	if err != nil {
		return err
	}

	// Process Secret input
	astSecret, err := builtins.StringOperand(args[1].Value, 2)
	if err != nil {
		return err
	}
	secret := string(astSecret)

	mac := hmac.New(sha512.New, []byte(secret))
	_, err = mac.Write([]byte(token.header + "." + token.payload))
	if err != nil {
		return err
	}

	signature, err := token.decodeSignature()
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(ast.Boolean(hmac.Equal([]byte(signature), mac.Sum(nil)))))
}

// -- Full JWT verification and decoding --

// Verification constraints. See tokens_test.go for unit tests.

// tokenConstraints holds decoded JWT verification constraints.
type tokenConstraints struct {
	// The set of asymmetric keys we can verify with.
	keys []interface{}

	// The single symmetric key we will verify with.
	secret string

	// The algorithm that must be used to verify.
	// If "", any algorithm is acceptable.
	alg string

	// The required issuer.
	// If "", any issuer is acceptable.
	iss string

	// The required audience.
	// If "", no audience is acceptable.
	aud string

	// The time to validate against, or -1 if no constraint set.
	// (If unset, the current time will be used.)
	time int64
}

// tokenConstraintHandler is the handler type for JWT verification constraints.
type tokenConstraintHandler func(value ast.Value, parameters *tokenConstraints) (err error)

// tokenConstraintTypes maps known JWT verification constraints to handlers.
var tokenConstraintTypes = map[string]tokenConstraintHandler{
	"cert": tokenConstraintCert,
	"secret": func(value ast.Value, constraints *tokenConstraints) (err error) {
		return tokenConstraintString("secret", value, &constraints.secret)
	},
	"alg": func(value ast.Value, constraints *tokenConstraints) (err error) {
		return tokenConstraintString("alg", value, &constraints.alg)
	},
	"iss": func(value ast.Value, constraints *tokenConstraints) (err error) {
		return tokenConstraintString("iss", value, &constraints.iss)
	},
	"aud": func(value ast.Value, constraints *tokenConstraints) (err error) {
		return tokenConstraintString("aud", value, &constraints.aud)
	},
	"time": tokenConstraintTime,
}

// tokenConstraintCert handles the `cert` constraint.
func tokenConstraintCert(value ast.Value, constraints *tokenConstraints) (err error) {
	var s ast.String
	var ok bool
	if s, ok = value.(ast.String); !ok {
		return fmt.Errorf("cert constraint: must be a string")
	}

	constraints.keys, err = getKeyFromCertOrJWK(string(s))
	return
}

// tokenConstraintTime handles the `time` constraint.
func tokenConstraintTime(value ast.Value, constraints *tokenConstraints) (err error) {
	var time ast.Number
	var ok bool
	if time, ok = value.(ast.Number); !ok {
		err = fmt.Errorf("token time constraint: must be a number")
		return
	}
	var timeFloat float64
	if timeFloat, err = strconv.ParseFloat(string(time), 64); err != nil {
		err = fmt.Errorf("token time constraint: %v", err)
		return
	}
	if timeFloat < 0 {
		err = fmt.Errorf("token time constraint: must not be negative")
		return
	}
	constraints.time = int64(timeFloat)
	return
}

// tokenConstraintString handles string constraints.
func tokenConstraintString(name string, value ast.Value, where *string) (err error) {
	var av ast.String
	var ok bool
	if av, ok = value.(ast.String); !ok {
		err = fmt.Errorf("%s constraint: must be a string", name)
		return
	}
	*where = string(av)
	return
}

// parseTokenConstraints parses the constraints argument.
func parseTokenConstraints(a ast.Value) (constraints tokenConstraints, err error) {
	constraints.time = -1
	var o ast.Object
	var ok bool
	if o, ok = a.(ast.Object); !ok {
		err = fmt.Errorf("token constraints must be object")
		return
	}
	if err = o.Iter(func(k *ast.Term, v *ast.Term) (err error) {
		var handler tokenConstraintHandler
		var ok bool
		name := string(k.Value.(ast.String))
		if handler, ok = tokenConstraintTypes[name]; ok {
			if err = handler(v.Value, &constraints); err != nil {
				return
			}
		} else {
			// Anything unknown is rejected.
			err = fmt.Errorf("unknown token validation constraint: %s", name)
			return
		}
		return
	}); err != nil {
		return
	}
	return
}

// validate validates the constraints argument.
func (constraints *tokenConstraints) validate() (err error) {
	keys := 0
	if constraints.keys != nil {
		keys++
	}
	if constraints.secret != "" {
		keys++
	}
	if keys > 1 {
		err = fmt.Errorf("duplicate key constraints")
		return
	}
	if keys < 1 {
		err = fmt.Errorf("no key constraint")
		return
	}
	return
}

// verify verifies a JWT using the constraints and the algorithm from the header
func (constraints *tokenConstraints) verify(kid, alg, header, payload, signature string) error {
	// Construct the payload
	plaintext := []byte(header)
	plaintext = append(plaintext, []byte(".")...)
	plaintext = append(plaintext, payload...)
	// Look up the algorithm
	var ok bool
	var a tokenAlgorithm
	a, ok = tokenAlgorithms[alg]
	if !ok {
		return fmt.Errorf("unknown JWS algorithm: %s", alg)
	}
	// If we're configured with asymmetric key(s) then only trust that
	if constraints.keys != nil {
		verified := false
		for _, key := range constraints.keys {
			err := a.verify(key, a.hash, plaintext, []byte(signature))
			if err == nil {
				verified = true
				break
			}
		}
		if !verified {
			return errSignatureNotVerified
		}
		return nil
	}
	if constraints.secret != "" {
		return a.verify([]byte(constraints.secret), a.hash, plaintext, []byte(signature))
	}
	// (*tokenConstraints)validate() should prevent this happening
	return errors.New("unexpectedly found no keys to trust")
}

// validAudience checks the audience of the JWT.
// It returns true if it meets the constraints and false otherwise.
func (constraints *tokenConstraints) validAudience(aud ast.Value) (valid bool) {
	var ok bool
	var s ast.String
	if s, ok = aud.(ast.String); ok {
		return string(s) == constraints.aud
	}
	var a *ast.Array
	if a, ok = aud.(*ast.Array); ok {
		for i := 0; i < a.Len(); i++ {
			if s, ok = a.Elem(i).Value.(ast.String); ok {
				if string(s) == constraints.aud {
					return true
				}
			} else {
				// Ill-formed aud claim
				return false
			}
		}
	}
	return false
}

// JWT algorithms

type tokenVerifyFunction func(key interface{}, hash crypto.Hash, payload []byte, signature []byte) (err error)
type tokenVerifyAsymmetricFunction func(key interface{}, hash crypto.Hash, digest []byte, signature []byte) (err error)

// jwtAlgorithm describes a JWS 'alg' value
type tokenAlgorithm struct {
	hash   crypto.Hash
	verify tokenVerifyFunction
}

// tokenAlgorithms is the known JWT algorithms
var tokenAlgorithms = map[string]tokenAlgorithm{
	"RS256": {crypto.SHA256, verifyAsymmetric(verifyRSAPKCS)},
	"RS384": {crypto.SHA384, verifyAsymmetric(verifyRSAPKCS)},
	"RS512": {crypto.SHA512, verifyAsymmetric(verifyRSAPKCS)},
	"PS256": {crypto.SHA256, verifyAsymmetric(verifyRSAPSS)},
	"PS384": {crypto.SHA384, verifyAsymmetric(verifyRSAPSS)},
	"PS512": {crypto.SHA512, verifyAsymmetric(verifyRSAPSS)},
	"ES256": {crypto.SHA256, verifyAsymmetric(verifyECDSA)},
	"ES384": {crypto.SHA384, verifyAsymmetric(verifyECDSA)},
	"ES512": {crypto.SHA512, verifyAsymmetric(verifyECDSA)},
	"HS256": {crypto.SHA256, verifyHMAC},
	"HS384": {crypto.SHA384, verifyHMAC},
	"HS512": {crypto.SHA512, verifyHMAC},
}

// errSignatureNotVerified is returned when a signature cannot be verified.
var errSignatureNotVerified = errors.New("signature not verified")

func verifyHMAC(key interface{}, hash crypto.Hash, payload []byte, signature []byte) (err error) {
	macKey, ok := key.([]byte)
	if !ok {
		return fmt.Errorf("incorrect symmetric key type")
	}
	mac := hmac.New(hash.New, macKey)
	if _, err = mac.Write([]byte(payload)); err != nil {
		return
	}
	if !hmac.Equal(signature, mac.Sum([]byte{})) {
		err = errSignatureNotVerified
	}
	return
}

func verifyAsymmetric(verify tokenVerifyAsymmetricFunction) tokenVerifyFunction {
	return func(key interface{}, hash crypto.Hash, payload []byte, signature []byte) (err error) {
		h := hash.New()
		h.Write(payload)
		return verify(key, hash, h.Sum([]byte{}), signature)
	}
}

func verifyRSAPKCS(key interface{}, hash crypto.Hash, digest []byte, signature []byte) (err error) {
	publicKeyRsa, ok := key.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("incorrect public key type")
	}
	if err = rsa.VerifyPKCS1v15(publicKeyRsa, hash, digest, signature); err != nil {
		err = errSignatureNotVerified
	}
	return
}

func verifyRSAPSS(key interface{}, hash crypto.Hash, digest []byte, signature []byte) (err error) {
	publicKeyRsa, ok := key.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("incorrect public key type")
	}
	if err = rsa.VerifyPSS(publicKeyRsa, hash, digest, signature, nil); err != nil {
		err = errSignatureNotVerified
	}
	return
}

func verifyECDSA(key interface{}, hash crypto.Hash, digest []byte, signature []byte) (err error) {
	publicKeyEcdsa, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("incorrect public key type")
	}
	r, s := &big.Int{}, &big.Int{}
	n := len(signature) / 2
	r.SetBytes(signature[:n])
	s.SetBytes(signature[n:])
	if ecdsa.Verify(publicKeyEcdsa, digest, r, s) {
		return nil
	}
	return errSignatureNotVerified
}

// JWT header parsing and parameters. See tokens_test.go for unit tests.

// tokenHeaderType represents a recognized JWT header field
// tokenHeader is a parsed JWT header
type tokenHeader struct {
	alg     string
	kid     string
	typ     string
	cty     string
	crit    map[string]bool
	unknown []string
}

// tokenHeaderHandler handles a JWT header parameters
type tokenHeaderHandler func(header *tokenHeader, value ast.Value) (err error)

// tokenHeaderTypes maps known JWT header parameters to handlers
var tokenHeaderTypes = map[string]tokenHeaderHandler{
	"alg": func(header *tokenHeader, value ast.Value) (err error) {
		return tokenHeaderString("alg", &header.alg, value)
	},
	"kid": func(header *tokenHeader, value ast.Value) (err error) {
		return tokenHeaderString("kid", &header.kid, value)
	},
	"typ": func(header *tokenHeader, value ast.Value) (err error) {
		return tokenHeaderString("typ", &header.typ, value)
	},
	"cty": func(header *tokenHeader, value ast.Value) (err error) {
		return tokenHeaderString("cty", &header.cty, value)
	},
	"crit": tokenHeaderCrit,
}

// tokenHeaderCrit handles the 'crit' header parameter
func tokenHeaderCrit(header *tokenHeader, value ast.Value) (err error) {
	var ok bool
	var v *ast.Array
	if v, ok = value.(*ast.Array); !ok {
		err = fmt.Errorf("crit: must be a list")
		return
	}
	header.crit = map[string]bool{}
	for i := 0; i < v.Len(); i++ {
		var tv ast.String
		if tv, ok = v.Elem(i).Value.(ast.String); !ok {
			err = fmt.Errorf("crit: must be a list of strings")
			return
		}
		header.crit[string(tv)] = true
	}
	if len(header.crit) == 0 {
		err = fmt.Errorf("crit: must be a nonempty list") // 'MUST NOT' use the empty list
		return
	}
	return
}

// tokenHeaderString handles string-format JWT header parameters
func tokenHeaderString(name string, where *string, value ast.Value) (err error) {
	var ok bool
	var v ast.String
	if v, ok = value.(ast.String); !ok {
		err = fmt.Errorf("%s: must be a string", name)
		return
	}
	*where = string(v)
	return
}

// parseTokenHeader parses the JWT header.
func parseTokenHeader(token *JSONWebToken) (header tokenHeader, err error) {
	header.unknown = []string{}
	if err = token.decodedHeader.Iter(func(k *ast.Term, v *ast.Term) (err error) {
		ks := string(k.Value.(ast.String))
		var ok bool
		var handler tokenHeaderHandler
		if handler, ok = tokenHeaderTypes[ks]; ok {
			if err = handler(&header, v.Value); err != nil {
				return
			}
		} else {
			header.unknown = append(header.unknown, ks)
		}
		return
	}); err != nil {
		return
	}
	return
}

// validTokenHeader returns true if the JOSE header is valid, otherwise false.
func (header *tokenHeader) valid() bool {
	// RFC7515 s4.1.1 alg MUST be present
	if header.alg == "" {
		return false
	}
	// RFC7515 4.1.11 JWS is invalid if there is a critical parameter that we did not recognize
	for _, u := range header.unknown {
		if header.crit[u] {
			return false
		}
	}
	return true
}

func commonBuiltinJWTEncodeSign(inputHeaders, jwsPayload, jwkSrc string) (v ast.Value, err error) {

	keys, err := jwk.ParseString(jwkSrc)
	if err != nil {
		return nil, err
	}
	key, err := keys.Keys[0].Materialize()
	if err != nil {
		return nil, err
	}
	if jwk.GetKeyTypeFromKey(key) != keys.Keys[0].GetKeyType() {
		return nil, fmt.Errorf("JWK derived key type and keyType parameter do not match")
	}

	standardHeaders := &jws.StandardHeaders{}
	jwsHeaders := []byte(inputHeaders)
	err = json.Unmarshal(jwsHeaders, standardHeaders)
	if err != nil {
		return nil, err
	}
	alg := standardHeaders.GetAlgorithm()

	if (standardHeaders.Type == "" || standardHeaders.Type == "JWT") && !json.Valid([]byte(jwsPayload)) {
		return nil, fmt.Errorf("type is JWT but payload is not JSON")
	}

	// process payload and sign
	var jwsCompact []byte
	jwsCompact, err = jws.SignLiteral([]byte(jwsPayload), alg, key, jwsHeaders)
	if err != nil {
		return nil, err
	}
	return ast.String(jwsCompact[:]), nil

}

func builtinJWTEncodeSign(a ast.Value, b ast.Value, c ast.Value) (v ast.Value, err error) {

	jwkSrc := c.String()

	inputHeaders := a.String()

	jwsPayload := b.String()

	return commonBuiltinJWTEncodeSign(inputHeaders, jwsPayload, jwkSrc)

}

func builtinJWTEncodeSignRaw(a ast.Value, b ast.Value, c ast.Value) (v ast.Value, err error) {

	jwkSrc, err := builtins.StringOperand(c, 1)
	if err != nil {
		return nil, err
	}
	inputHeaders, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	jwsPayload, err := builtins.StringOperand(b, 1)
	if err != nil {
		return nil, err
	}
	return commonBuiltinJWTEncodeSign(string(inputHeaders), string(jwsPayload), string(jwkSrc))
}

// Implements full JWT decoding, validation and verification.
func builtinJWTDecodeVerify(a ast.Value, b ast.Value) (v ast.Value, err error) {
	// io.jwt.decode_verify(string, constraints, [valid, header, payload])
	//
	// If valid is true then the signature verifies and all constraints are met.
	// If valid is false then either the signature did not verify or some constrain
	// was not met.
	//
	// Decoding errors etc are returned as errors.
	arr := []*ast.Term{
		ast.BooleanTerm(false), // by default, not verified
		ast.NewTerm(ast.NewObject()),
		ast.NewTerm(ast.NewObject()),
	}
	var constraints tokenConstraints
	if constraints, err = parseTokenConstraints(b); err != nil {
		return
	}
	if err = constraints.validate(); err != nil {
		return
	}
	var token *JSONWebToken
	var p ast.Value
	for {
		// RFC7519 7.2 #1-2 split into parts
		if token, err = decodeJWT(a); err != nil {
			return
		}
		// RFC7519 7.2 #3, #4, #6
		if err = token.decodeHeader(); err != nil {
			return
		}
		// RFC7159 7.2 #5 (and RFC7159 5.2 #5) validate header fields
		var header tokenHeader
		if header, err = parseTokenHeader(token); err != nil {
			return
		}
		if !header.valid() {
			return ast.NewArray(arr...), nil
		}
		// Check constraints that impact signature verification.
		if constraints.alg != "" && constraints.alg != header.alg {
			return ast.NewArray(arr...), nil
		}
		// RFC7159 7.2 #7 verify the signature
		var signature string
		if signature, err = token.decodeSignature(); err != nil {
			return
		}
		if err = constraints.verify(header.kid, header.alg, token.header, token.payload, signature); err != nil {
			if err == errSignatureNotVerified {
				return ast.NewArray(arr...), nil
			}
			return
		}
		// RFC7159 7.2 #9-10 decode the payload
		if p, err = builtinBase64UrlDecode(ast.String(token.payload)); err != nil {
			return nil, fmt.Errorf("JWT payload had invalid encoding: %v", err)
		}
		// RFC7159 7.2 #8 and 5.2 cty
		if strings.ToUpper(header.cty) == "JWT" {
			// Nested JWT, go round again
			a = p
			continue
		} else {
			// Non-nested JWT (or we've reached the bottom of the nesting).
			break
		}
	}
	var payload ast.Object
	if payload, err = extractJSONObject(string(p.(ast.String))); err != nil {
		return
	}
	// Check registered claim names against constraints or environment
	// RFC7159 4.1.1 iss
	if constraints.iss != "" {
		if iss := payload.Get(jwtIssKey); iss != nil {
			issVal := string(iss.Value.(ast.String))
			if constraints.iss != issVal {
				return ast.NewArray(arr...), nil
			}
		}
	}
	// RFC7159 4.1.3 aud
	if aud := payload.Get(jwtAudKey); aud != nil {
		if !constraints.validAudience(aud.Value) {
			return ast.NewArray(arr...), nil
		}
	} else {
		if constraints.aud != "" {
			return ast.NewArray(arr...), nil
		}
	}
	// RFC7159 4.1.4 exp
	if exp := payload.Get(jwtExpKey); exp != nil {
		if constraints.time < 0 {
			constraints.time = time.Now().UnixNano()
		}

		// constraints.time is in nanoseconds but exp Value is in seconds
		compareTime := ast.Number(strconv.FormatFloat(float64(constraints.time)/1000000000, 'g', -1, 64))

		if ast.Compare(compareTime, exp.Value.(ast.Number)) != -1 {
			return ast.NewArray(arr...), nil
		}
	}
	// RFC7159 4.1.5 nbf
	if nbf := payload.Get(jwtNbfKey); nbf != nil {
		if constraints.time < 0 {
			constraints.time = time.Now().UnixNano()
		}

		// constraints.time is in nanoseconds but nbf Value is in seconds
		compareTime := ast.Number(strconv.FormatFloat(float64(constraints.time)/1000000000, 'g', -1, 64))

		if ast.Compare(compareTime, nbf.Value.(ast.Number)) == -1 {
			return ast.NewArray(arr...), nil
		}
	}
	// Format the result
	arr[0] = ast.BooleanTerm(true)
	arr[1] = ast.NewTerm(token.decodedHeader)
	arr[2] = ast.NewTerm(payload)
	return ast.NewArray(arr...), nil
}

// -- Utilities --

func decodeJWT(a ast.Value) (*JSONWebToken, error) {
	// Parse the JSON Web Token
	astEncode, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	encoding := string(astEncode)
	if !strings.Contains(encoding, ".") {
		return nil, errors.New("encoded JWT had no period separators")
	}

	parts := strings.Split(encoding, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("encoded JWT must have 3 sections, found %d", len(parts))
	}

	return &JSONWebToken{header: parts[0], payload: parts[1], signature: parts[2]}, nil
}

func (token *JSONWebToken) decodeSignature() (string, error) {
	decodedSignature, err := builtinBase64UrlDecode(ast.String(token.signature))
	if err != nil {
		return "", err
	}

	signatureAst, err := builtins.StringOperand(decodedSignature, 1)
	if err != nil {
		return "", err
	}
	return string(signatureAst), err
}

// Extract, validate and return the JWT header as an ast.Object.
func validateJWTHeader(h string) (ast.Object, error) {
	header, err := extractJSONObject(h)
	if err != nil {
		return nil, fmt.Errorf("bad JWT header: %v", err)
	}

	// There are two kinds of JWT tokens, a JSON Web Signature (JWS) and
	// a JSON Web Encryption (JWE). The latter is very involved, and we
	// won't support it for now.
	// This code checks which kind of JWT we are dealing with according to
	// RFC 7516 Section 9: https://tools.ietf.org/html/rfc7516#section-9
	if header.Get(jwtEncKey) != nil {
		return nil, errors.New("JWT is a JWE object, which is not supported")
	}

	return header, nil
}

func extractJSONObject(s string) (ast.Object, error) {
	// XXX: This code relies on undocumented behavior of Go's
	// json.Unmarshal using the last occurrence of duplicate keys in a JSON
	// Object. If duplicate keys are present in a JWT, the last must be
	// used or the token rejected. Since detecting duplicates is tantamount
	// to parsing it ourselves, we're relying on the Go implementation
	// using the last occurring instance of the key, which is the behavior
	// as of Go 1.8.1.
	v, err := builtinJSONUnmarshal(ast.String(s))
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	o, ok := v.(ast.Object)
	if !ok {
		return nil, errors.New("decoded JSON type was not an Object")
	}

	return o, nil
}

// getInputSha returns the SHA checksum of the input
func getInputSHA(input []byte, h func() hash.Hash) (hash []byte) {
	hasher := h()
	hasher.Write(input)
	return hasher.Sum(nil)
}

func init() {
	RegisterFunctionalBuiltin1(ast.JWTDecode.Name, builtinJWTDecode)
	RegisterBuiltinFunc(ast.JWTVerifyRS256.Name, builtinJWTVerifyRS256)
	RegisterBuiltinFunc(ast.JWTVerifyRS384.Name, builtinJWTVerifyRS384)
	RegisterBuiltinFunc(ast.JWTVerifyRS512.Name, builtinJWTVerifyRS512)
	RegisterBuiltinFunc(ast.JWTVerifyPS256.Name, builtinJWTVerifyPS256)
	RegisterBuiltinFunc(ast.JWTVerifyPS384.Name, builtinJWTVerifyPS384)
	RegisterBuiltinFunc(ast.JWTVerifyPS512.Name, builtinJWTVerifyPS512)
	RegisterBuiltinFunc(ast.JWTVerifyES256.Name, builtinJWTVerifyES256)
	RegisterBuiltinFunc(ast.JWTVerifyES384.Name, builtinJWTVerifyES384)
	RegisterBuiltinFunc(ast.JWTVerifyES512.Name, builtinJWTVerifyES512)
	RegisterBuiltinFunc(ast.JWTVerifyHS256.Name, builtinJWTVerifyHS256)
	RegisterBuiltinFunc(ast.JWTVerifyHS384.Name, builtinJWTVerifyHS384)
	RegisterBuiltinFunc(ast.JWTVerifyHS512.Name, builtinJWTVerifyHS512)
	RegisterFunctionalBuiltin2(ast.JWTDecodeVerify.Name, builtinJWTDecodeVerify)
	RegisterFunctionalBuiltin3(ast.JWTEncodeSignRaw.Name, builtinJWTEncodeSignRaw)
	RegisterFunctionalBuiltin3(ast.JWTEncodeSign.Name, builtinJWTEncodeSign)
}
