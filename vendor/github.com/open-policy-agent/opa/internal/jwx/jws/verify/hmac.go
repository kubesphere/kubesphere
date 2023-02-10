package verify

import (
	"crypto/hmac"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/internal/jwx/jwa"
	"github.com/open-policy-agent/opa/internal/jwx/jws/sign"
)

func newHMAC(alg jwa.SignatureAlgorithm) (*HMACVerifier, error) {

	s, err := sign.New(alg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC signer: %w", err)
	}
	return &HMACVerifier{signer: s}, nil
}

// Verify checks whether the signature for a given input and key is correct
func (v HMACVerifier) Verify(signingInput, signature []byte, key interface{}) (err error) {

	expected, err := v.signer.Sign(signingInput, key)
	if err != nil {
		return fmt.Errorf("failed to generated signature: %w", err)
	}

	if !hmac.Equal(signature, expected) {
		return errors.New("failed to match hmac signature")
	}
	return nil
}
