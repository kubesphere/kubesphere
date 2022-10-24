package jwk

import (
	"fmt"

	"github.com/open-policy-agent/opa/internal/jwx/jwa"
)

func newSymmetricKey(key []byte) (*SymmetricKey, error) {
	var hdr StandardHeaders

	err := hdr.Set(KeyTypeKey, jwa.OctetSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to set Key Type: %w", err)
	}
	return &SymmetricKey{
		StandardHeaders: &hdr,
		key:             key,
	}, nil
}

// Materialize returns the octets for this symmetric key.
// Since this is a symmetric key, this just calls Octets
func (s SymmetricKey) Materialize() (interface{}, error) {
	return s.Octets(), nil
}

// Octets returns the octets in the key
func (s SymmetricKey) Octets() []byte {
	return s.key
}

// GenerateKey creates a Symmetric key from a RawKeyJSON
func (s *SymmetricKey) GenerateKey(keyJSON *RawKeyJSON) error {

	*s = SymmetricKey{
		StandardHeaders: &keyJSON.StandardHeaders,
		key:             keyJSON.K,
	}
	return nil
}
