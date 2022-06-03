package jwk

import (
	"crypto/rsa"
	"math/big"

	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/internal/jwx/jwa"
)

func newRSAPublicKey(key *rsa.PublicKey) (*RSAPublicKey, error) {

	var hdr StandardHeaders
	err := hdr.Set(KeyTypeKey, jwa.RSA)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to set Key Type")
	}
	return &RSAPublicKey{
		StandardHeaders: &hdr,
		key:             key,
	}, nil
}

func newRSAPrivateKey(key *rsa.PrivateKey) (*RSAPrivateKey, error) {

	var hdr StandardHeaders
	err := hdr.Set(KeyTypeKey, jwa.RSA)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to set Key Type")
	}
	return &RSAPrivateKey{
		StandardHeaders: &hdr,
		key:             key,
	}, nil
}

// Materialize returns the standard RSA Public Key representation stored in the internal representation
func (k *RSAPublicKey) Materialize() (interface{}, error) {
	if k.key == nil {
		return nil, errors.New(`key has no rsa.PublicKey associated with it`)
	}
	return k.key, nil
}

// Materialize returns the standard RSA Private Key representation stored in the internal representation
func (k *RSAPrivateKey) Materialize() (interface{}, error) {
	if k.key == nil {
		return nil, errors.New(`key has no rsa.PrivateKey associated with it`)
	}
	return k.key, nil
}

// GenerateKey creates a RSAPublicKey from a RawKeyJSON
func (k *RSAPublicKey) GenerateKey(keyJSON *RawKeyJSON) error {

	if keyJSON.N == nil || keyJSON.E == nil {
		return errors.Errorf("Missing mandatory key parameters N or E")
	}
	rsaPublicKey := &rsa.PublicKey{
		N: (&big.Int{}).SetBytes(keyJSON.N.Bytes()),
		E: int((&big.Int{}).SetBytes(keyJSON.E.Bytes()).Int64()),
	}
	k.key = rsaPublicKey
	k.StandardHeaders = &keyJSON.StandardHeaders
	return nil
}

// GenerateKey creates a RSAPublicKey from a RawKeyJSON
func (k *RSAPrivateKey) GenerateKey(keyJSON *RawKeyJSON) error {

	rsaPublicKey := &RSAPublicKey{}
	err := rsaPublicKey.GenerateKey(keyJSON)
	if err != nil {
		return errors.Wrap(err, "failed to generate public key")
	}

	if keyJSON.D == nil || keyJSON.P == nil || keyJSON.Q == nil {
		return errors.Errorf("Missing mandatory key parameters D, P or Q")
	}
	privateKey := &rsa.PrivateKey{
		PublicKey: *rsaPublicKey.key,
		D:         (&big.Int{}).SetBytes(keyJSON.D.Bytes()),
		Primes: []*big.Int{
			(&big.Int{}).SetBytes(keyJSON.P.Bytes()),
			(&big.Int{}).SetBytes(keyJSON.Q.Bytes()),
		},
	}

	if keyJSON.Dp.Len() > 0 {
		privateKey.Precomputed.Dp = (&big.Int{}).SetBytes(keyJSON.Dp.Bytes())
	}
	if keyJSON.Dq.Len() > 0 {
		privateKey.Precomputed.Dq = (&big.Int{}).SetBytes(keyJSON.Dq.Bytes())
	}
	if keyJSON.Qi.Len() > 0 {
		privateKey.Precomputed.Qinv = (&big.Int{}).SetBytes(keyJSON.Qi.Bytes())
	}

	k.key = privateKey
	k.StandardHeaders = &keyJSON.StandardHeaders
	return nil
}
