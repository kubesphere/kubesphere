// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package bundle provide helpers that assist in creating the verification and signing key configuration
package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/open-policy-agent/opa/internal/jwx/jwa"
	"github.com/open-policy-agent/opa/internal/jwx/jws/sign"

	"github.com/open-policy-agent/opa/util"
)

const (
	defaultTokenSigningAlg = "RS256"
)

// VerificationConfig represents the key configuration used to verify a signed bundle
type VerificationConfig struct {
	PublicKeys map[string]*KeyConfig
	KeyID      string   `json:"keyid"`
	Scope      string   `json:"scope"`
	Exclude    []string `json:"exclude_files"`
}

// NewVerificationConfig return a new VerificationConfig
func NewVerificationConfig(keys map[string]*KeyConfig, id, scope string, exclude []string) *VerificationConfig {
	return &VerificationConfig{
		PublicKeys: keys,
		KeyID:      id,
		Scope:      scope,
		Exclude:    exclude,
	}
}

// ValidateAndInjectDefaults validates the config and inserts default values
func (vc *VerificationConfig) ValidateAndInjectDefaults(keys map[string]*KeyConfig) error {
	vc.PublicKeys = keys

	if vc.KeyID != "" {
		found := false
		for key := range keys {
			if key == vc.KeyID {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("key id %s not found", vc.KeyID)
		}
	}
	return nil
}

// GetPublicKey returns the public key corresponding to the given key id
func (vc *VerificationConfig) GetPublicKey(id string) (*KeyConfig, error) {
	var kc *KeyConfig
	var ok bool

	if kc, ok = vc.PublicKeys[id]; !ok {
		return nil, fmt.Errorf("verification key corresponding to ID %v not found", id)
	}
	return kc, nil
}

// KeyConfig holds the actual public keys used to verify a signed bundle
type KeyConfig struct {
	Key       string `json:"key"`
	Algorithm string `json:"algorithm"`
	Scope     string `json:"scope"`
}

// NewKeyConfig return a new KeyConfig
func NewKeyConfig(key, alg, scope string) *KeyConfig {
	return &KeyConfig{
		Key:       key,
		Algorithm: alg,
		Scope:     scope,
	}
}

// ParseKeysConfig returns a map containing the public key and the signing algorithm
func ParseKeysConfig(raw json.RawMessage) (map[string]*KeyConfig, error) {
	keys := map[string]*KeyConfig{}
	var obj map[string]json.RawMessage

	if err := util.Unmarshal(raw, &obj); err == nil {
		for k := range obj {
			var keyConfig KeyConfig
			if err = util.Unmarshal(obj[k], &keyConfig); err != nil {
				return nil, err
			}

			if err = keyConfig.validateAndInjectDefaults(k); err != nil {
				return nil, err
			}

			keys[k] = &keyConfig
		}
	} else {
		return nil, err
	}
	return keys, nil
}

func (k *KeyConfig) validateAndInjectDefaults(id string) error {
	if k.Key == "" {
		return fmt.Errorf("invalid keys configuration: verification key empty for key ID %v", id)
	}

	if k.Algorithm == "" {
		k.Algorithm = defaultTokenSigningAlg
	}

	return nil
}

// Equal returns true if this key config is equal to the other.
func (k *KeyConfig) Equal(other *KeyConfig) bool {
	return reflect.DeepEqual(k, other)
}

// SigningConfig represents the key configuration used to generate a signed bundle
type SigningConfig struct {
	Key        string
	Algorithm  string
	ClaimsPath string
}

// NewSigningConfig return a new SigningConfig
func NewSigningConfig(key, alg, claimsPath string) *SigningConfig {
	if alg == "" {
		alg = defaultTokenSigningAlg
	}

	return &SigningConfig{
		Key:        key,
		Algorithm:  alg,
		ClaimsPath: claimsPath,
	}
}

// GetPrivateKey returns the private key or secret from the signing config
func (s *SigningConfig) GetPrivateKey() (interface{}, error) {
	var priv string
	if _, err := os.Stat(s.Key); err == nil {
		bs, err := ioutil.ReadFile(s.Key)
		if err != nil {
			return nil, err
		}
		priv = string(bs)
	} else if os.IsNotExist(err) {
		priv = s.Key
	} else {
		return nil, err
	}
	return sign.GetSigningKey(priv, jwa.SignatureAlgorithm(s.Algorithm))
}

// GetClaims returns the claims by reading the file specified in the signing config
func (s *SigningConfig) GetClaims() (map[string]interface{}, error) {
	var claims map[string]interface{}

	bs, err := ioutil.ReadFile(s.ClaimsPath)
	if err != nil {
		return claims, err
	}

	if err := util.UnmarshalJSON(bs, &claims); err != nil {
		return claims, err
	}
	return claims, nil
}
