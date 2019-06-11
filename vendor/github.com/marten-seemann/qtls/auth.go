// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qtls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"hash"
	"io"
)

// pickSignatureAlgorithm selects a signature algorithm that is compatible with
// the given public key and the list of algorithms from the peer and this side.
// The lists of signature algorithms (peerSigAlgs and ourSigAlgs) are ignored
// for tlsVersion < VersionTLS12.
//
// The returned SignatureScheme codepoint is only meaningful for TLS 1.2,
// previous TLS versions have a fixed hash function.
func pickSignatureAlgorithm(pubkey crypto.PublicKey, peerSigAlgs, ourSigAlgs []SignatureScheme, tlsVersion uint16) (sigAlg SignatureScheme, sigType uint8, hashFunc crypto.Hash, err error) {
	if tlsVersion < VersionTLS12 || len(peerSigAlgs) == 0 {
		// For TLS 1.1 and before, the signature algorithm could not be
		// negotiated and the hash is fixed based on the signature type. For TLS
		// 1.2, if the client didn't send signature_algorithms extension then we
		// can assume that it supports SHA1. See RFC 5246, Section 7.4.1.4.1.
		switch pubkey.(type) {
		case *rsa.PublicKey:
			if tlsVersion < VersionTLS12 {
				return 0, signaturePKCS1v15, crypto.MD5SHA1, nil
			} else {
				return PKCS1WithSHA1, signaturePKCS1v15, crypto.SHA1, nil
			}
		case *ecdsa.PublicKey:
			return ECDSAWithSHA1, signatureECDSA, crypto.SHA1, nil
		default:
			return 0, 0, 0, fmt.Errorf("tls: unsupported public key: %T", pubkey)
		}
	}
	for _, sigAlg := range peerSigAlgs {
		if !isSupportedSignatureAlgorithm(sigAlg, ourSigAlgs) {
			continue
		}
		hashAlg, err := hashFromSignatureScheme(sigAlg)
		if err != nil {
			panic("tls: supported signature algorithm has an unknown hash function")
		}
		sigType := signatureFromSignatureScheme(sigAlg)
		switch pubkey.(type) {
		case *rsa.PublicKey:
			if sigType == signaturePKCS1v15 || sigType == signatureRSAPSS {
				return sigAlg, sigType, hashAlg, nil
			}
		case *ecdsa.PublicKey:
			if sigType == signatureECDSA {
				return sigAlg, sigType, hashAlg, nil
			}
		default:
			return 0, 0, 0, fmt.Errorf("tls: unsupported public key: %T", pubkey)
		}
	}
	return 0, 0, 0, errors.New("tls: peer doesn't support any common signature algorithms")
}

// verifyHandshakeSignature verifies a signature against pre-hashed handshake
// contents.
func verifyHandshakeSignature(sigType uint8, pubkey crypto.PublicKey, hashFunc crypto.Hash, digest, sig []byte) error {
	switch sigType {
	case signatureECDSA:
		pubKey, ok := pubkey.(*ecdsa.PublicKey)
		if !ok {
			return errors.New("tls: ECDSA signing requires a ECDSA public key")
		}
		ecdsaSig := new(ecdsaSignature)
		if _, err := asn1.Unmarshal(sig, ecdsaSig); err != nil {
			return err
		}
		if ecdsaSig.R.Sign() <= 0 || ecdsaSig.S.Sign() <= 0 {
			return errors.New("tls: ECDSA signature contained zero or negative values")
		}
		if !ecdsa.Verify(pubKey, digest, ecdsaSig.R, ecdsaSig.S) {
			return errors.New("tls: ECDSA verification failure")
		}
	case signaturePKCS1v15:
		pubKey, ok := pubkey.(*rsa.PublicKey)
		if !ok {
			return errors.New("tls: RSA signing requires a RSA public key")
		}
		if err := rsa.VerifyPKCS1v15(pubKey, hashFunc, digest, sig); err != nil {
			return err
		}
	case signatureRSAPSS:
		pubKey, ok := pubkey.(*rsa.PublicKey)
		if !ok {
			return errors.New("tls: RSA signing requires a RSA public key")
		}
		signOpts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}
		if err := rsa.VerifyPSS(pubKey, hashFunc, digest, sig, signOpts); err != nil {
			return err
		}
	default:
		return errors.New("tls: unknown signature algorithm")
	}
	return nil
}

const (
	serverSignatureContext = "TLS 1.3, server CertificateVerify\x00"
	clientSignatureContext = "TLS 1.3, client CertificateVerify\x00"
)

var signaturePadding = []byte{
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
}

// writeSignedMessage writes the content to be signed by certificate keys in TLS
// 1.3 to sigHash. See RFC 8446, Section 4.4.3.
func writeSignedMessage(sigHash io.Writer, context string, transcript hash.Hash) {
	sigHash.Write(signaturePadding)
	io.WriteString(sigHash, context)
	sigHash.Write(transcript.Sum(nil))
}

// signatureSchemesForCertificate returns the list of supported SignatureSchemes
// for a given certificate, based on the public key and the protocol version. It
// does not support the crypto.Decrypter interface, so shouldn't be used on the
// server side in TLS 1.2 and earlier.
func signatureSchemesForCertificate(version uint16, cert *Certificate) []SignatureScheme {
	priv, ok := cert.PrivateKey.(crypto.Signer)
	if !ok {
		return nil
	}

	switch pub := priv.Public().(type) {
	case *ecdsa.PublicKey:
		if version != VersionTLS13 {
			// In TLS 1.2 and earlier, ECDSA algorithms are not
			// constrained to a single curve.
			return []SignatureScheme{
				ECDSAWithP256AndSHA256,
				ECDSAWithP384AndSHA384,
				ECDSAWithP521AndSHA512,
				ECDSAWithSHA1,
			}
		}
		switch pub.Curve {
		case elliptic.P256():
			return []SignatureScheme{ECDSAWithP256AndSHA256}
		case elliptic.P384():
			return []SignatureScheme{ECDSAWithP384AndSHA384}
		case elliptic.P521():
			return []SignatureScheme{ECDSAWithP521AndSHA512}
		default:
			return nil
		}
	case *rsa.PublicKey:
		if version != VersionTLS13 {
			return []SignatureScheme{
				PSSWithSHA256,
				PSSWithSHA384,
				PSSWithSHA512,
				PKCS1WithSHA256,
				PKCS1WithSHA384,
				PKCS1WithSHA512,
				PKCS1WithSHA1,
			}
		}
		// RSA keys with RSA-PSS OID are not supported by crypto/x509.
		return []SignatureScheme{
			PSSWithSHA256,
			PSSWithSHA384,
			PSSWithSHA512,
		}
	default:
		return nil
	}
}

// unsupportedCertificateError returns a helpful error for certificates with
// an unsupported private key.
func unsupportedCertificateError(cert *Certificate) error {
	switch cert.PrivateKey.(type) {
	case rsa.PrivateKey, ecdsa.PrivateKey:
		return fmt.Errorf("tls: unsupported certificate: private key is %T, expected *%T",
			cert.PrivateKey, cert.PrivateKey)
	}

	signer, ok := cert.PrivateKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("tls: certificate private key (%T) does not implement crypto.Signer",
			cert.PrivateKey)
	}

	switch pub := signer.Public().(type) {
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P256():
		case elliptic.P384():
		case elliptic.P521():
		default:
			return fmt.Errorf("tls: unsupported certificate curve (%s)", pub.Curve.Params().Name)
		}
	case *rsa.PublicKey:
	default:
		return fmt.Errorf("tls: unsupported certificate key (%T)", pub)
	}

	return fmt.Errorf("tls: internal error: unsupported key (%T)", cert.PrivateKey)
}
