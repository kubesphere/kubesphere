// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"os"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/internal/jwx/jwk"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/util"
)

const (
	// blockTypeCertificate indicates this PEM block contains the signed certificate.
	// Exported for tests.
	blockTypeCertificate = "CERTIFICATE"
	// blockTypeCertificateRequest indicates this PEM block contains a certificate
	// request. Exported for tests.
	blockTypeCertificateRequest = "CERTIFICATE REQUEST"
	// blockTypeRSAPrivateKey indicates this PEM block contains a RSA private key.
	// Exported for tests.
	blockTypeRSAPrivateKey = "RSA PRIVATE KEY"
	// blockTypeRSAPrivateKey indicates this PEM block contains a RSA private key.
	// Exported for tests.
	blockTypePrivateKey   = "PRIVATE KEY"
	blockTypeEcPrivateKey = "EC PRIVATE KEY"
)

func builtinCryptoX509ParseCertificates(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	input, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	certs, err := getX509CertsFromString(string(input))
	if err != nil {
		return err
	}

	v, err := ast.InterfaceToValue(extendCertificates(certs))
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(v))
}

// extendedCert is a wrapper around x509.Certificate that adds additional fields for JSON serialization.
type extendedCert struct {
	x509.Certificate
	URIStrings []string
}

func extendCertificates(certs []*x509.Certificate) []extendedCert {
	// add a field to certs containing the URIs as strings
	processedCerts := make([]extendedCert, len(certs))

	for i, cert := range certs {
		processedCerts[i].Certificate = *cert
		if cert.URIs != nil {
			processedCerts[i].URIStrings = make([]string, len(cert.URIs))
			for j, uri := range cert.URIs {
				processedCerts[i].URIStrings[j] = uri.String()
			}
		}
	}
	return processedCerts
}

func builtinCryptoX509ParseAndVerifyCertificates(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	a := operands[0].Value
	input, err := builtins.StringOperand(a, 1)
	if err != nil {
		return err
	}

	invalid := ast.ArrayTerm(
		ast.BooleanTerm(false),
		ast.NewTerm(ast.NewArray()),
	)

	certs, err := getX509CertsFromString(string(input))
	if err != nil {
		return iter(invalid)
	}

	verified, err := verifyX509CertificateChain(certs, x509.VerifyOptions{})
	if err != nil {
		return iter(invalid)
	}

	value, err := ast.InterfaceToValue(extendCertificates(verified))
	if err != nil {
		return err
	}

	valid := ast.ArrayTerm(
		ast.BooleanTerm(true),
		ast.NewTerm(value),
	)

	return iter(valid)
}

var allowedKeyUsages = map[string]x509.ExtKeyUsage{
	"KeyUsageAny":                            x509.ExtKeyUsageAny,
	"KeyUsageServerAuth":                     x509.ExtKeyUsageServerAuth,
	"KeyUsageClientAuth":                     x509.ExtKeyUsageClientAuth,
	"KeyUsageCodeSigning":                    x509.ExtKeyUsageCodeSigning,
	"KeyUsageEmailProtection":                x509.ExtKeyUsageEmailProtection,
	"KeyUsageIPSECEndSystem":                 x509.ExtKeyUsageIPSECEndSystem,
	"KeyUsageIPSECTunnel":                    x509.ExtKeyUsageIPSECTunnel,
	"KeyUsageIPSECUser":                      x509.ExtKeyUsageIPSECUser,
	"KeyUsageTimeStamping":                   x509.ExtKeyUsageTimeStamping,
	"KeyUsageOCSPSigning":                    x509.ExtKeyUsageOCSPSigning,
	"KeyUsageMicrosoftServerGatedCrypto":     x509.ExtKeyUsageMicrosoftServerGatedCrypto,
	"KeyUsageNetscapeServerGatedCrypto":      x509.ExtKeyUsageNetscapeServerGatedCrypto,
	"KeyUsageMicrosoftCommercialCodeSigning": x509.ExtKeyUsageMicrosoftCommercialCodeSigning,
	"KeyUsageMicrosoftKernelCodeSigning":     x509.ExtKeyUsageMicrosoftKernelCodeSigning,
}

func builtinCryptoX509ParseAndVerifyCertificatesWithOptions(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	input, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	options, err := builtins.ObjectOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	invalid := ast.ArrayTerm(
		ast.BooleanTerm(false),
		ast.NewTerm(ast.NewArray()),
	)

	certs, err := getX509CertsFromString(string(input))
	if err != nil {
		return iter(invalid)
	}

	// Collect the cert verification options
	verifyOpt, err := extractVerifyOpts(options)
	if err != nil {
		return err
	}

	verified, err := verifyX509CertificateChain(certs, verifyOpt)
	if err != nil {
		return iter(invalid)
	}

	value, err := ast.InterfaceToValue(verified)
	if err != nil {
		return err
	}

	valid := ast.ArrayTerm(
		ast.BooleanTerm(true),
		ast.NewTerm(value),
	)

	return iter(valid)
}

func extractVerifyOpts(options ast.Object) (verifyOpt x509.VerifyOptions, err error) {

	for _, key := range options.Keys() {
		k, err := ast.JSON(key.Value)
		if err != nil {
			return verifyOpt, err
		}
		k, ok := k.(string)
		if !ok {
			continue
		}

		switch k {
		case "DNSName":
			dns, ok := options.Get(key).Value.(ast.String)
			if ok {
				verifyOpt.DNSName = strings.Trim(string(dns), "\"")
			} else {
				return verifyOpt, fmt.Errorf("'DNSName' should be a string")
			}
		case "CurrentTime":
			c, ok := options.Get(key).Value.(ast.Number)
			if ok {
				nanosecs, ok := c.Int64()
				if ok {
					verifyOpt.CurrentTime = time.Unix(0, nanosecs)
				} else {
					return verifyOpt, fmt.Errorf("'CurrentTime' should be a valid int64 number")
				}
			} else {
				return verifyOpt, fmt.Errorf("'CurrentTime' should be a number")
			}
		case "MaxConstraintComparisons":
			c, ok := options.Get(key).Value.(ast.Number)
			if ok {
				maxComparisons, ok := c.Int()
				if ok {
					verifyOpt.MaxConstraintComparisions = maxComparisons
				} else {
					return verifyOpt, fmt.Errorf("'MaxConstraintComparisons' should be a valid number")
				}
			} else {
				return verifyOpt, fmt.Errorf("'MaxConstraintComparisons' should be a number")
			}
		case "KeyUsages":
			type forEach interface {
				Foreach(func(*ast.Term))
			}
			var ks forEach
			switch options.Get(key).Value.(type) {
			case *ast.Array:
				ks = options.Get(key).Value.(*ast.Array)
			case ast.Set:
				ks = options.Get(key).Value.(ast.Set)
			default:
				return verifyOpt, fmt.Errorf("'KeyUsages' should be an Array or Set")
			}

			// Collect the x509.ExtKeyUsage values by looking up the
			// mapping of key usage strings to x509.ExtKeyUsage
			var invalidKUsgs []string
			ks.Foreach(func(t *ast.Term) {
				u, ok := t.Value.(ast.String)
				if ok {
					v := strings.Trim(string(u), "\"")
					if k, ok := allowedKeyUsages[v]; ok {
						verifyOpt.KeyUsages = append(verifyOpt.KeyUsages, k)
					} else {
						invalidKUsgs = append(invalidKUsgs, v)
					}
				}
			})
			if len(invalidKUsgs) > 0 {
				return x509.VerifyOptions{}, fmt.Errorf("invalid entries for 'KeyUsages' found: %s", invalidKUsgs)
			}
		default:
			return verifyOpt, fmt.Errorf("invalid key option")
		}

	}

	return verifyOpt, nil
}

func builtinCryptoX509ParseKeyPair(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	certificate, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	key, err := builtins.StringOperand(operands[1].Value, 1)
	if err != nil {
		return err
	}

	certs, err := getTLSx509KeyPairFromString([]byte(certificate), []byte(key))
	if err != nil {
		return err
	}
	v, err := ast.InterfaceToValue(certs)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(v))
}

func builtinCryptoX509ParseCertificateRequest(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	input, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// data to be passed to x509.ParseCertificateRequest
	bytes := []byte(input)

	// if the input is not a PEM string, attempt to decode b64
	if str := string(input); !strings.HasPrefix(str, "-----BEGIN CERTIFICATE REQUEST-----") {
		bytes, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}
	}

	p, _ := pem.Decode(bytes)
	if p != nil && p.Type != blockTypeCertificateRequest {
		return fmt.Errorf("invalid PEM-encoded certificate signing request")
	}
	if p != nil {
		bytes = p.Bytes
	}

	csr, err := x509.ParseCertificateRequest(bytes)
	if err != nil {
		return err
	}

	bs, err := json.Marshal(csr)
	if err != nil {
		return err
	}

	var x interface{}
	if err := util.UnmarshalJSON(bs, &x); err != nil {
		return err
	}

	v, err := ast.InterfaceToValue(x)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(v))
}

func builtinCryptoJWKFromPrivateKey(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var x interface{}

	a := operands[0].Value
	input, err := builtins.StringOperand(a, 1)
	if err != nil {
		return err
	}

	// get the raw private key
	pemDataString := string(input)

	if pemDataString == "" {
		return fmt.Errorf("input PEM data was empty")
	}

	// This built in must be supplied a valid PEM or base64 encoded string.
	// If the input is not a PEM string, attempt to decode b64.
	// If the base64 decode fails - this is an error
	if !strings.HasPrefix(pemDataString, "-----BEGIN") {
		bs, err := base64.StdEncoding.DecodeString(pemDataString)
		if err != nil {
			return err
		}
		pemDataString = string(bs)
	}

	rawKeys, err := getPrivateKeysFromPEMData(pemDataString)
	if err != nil {
		return err
	}

	if len(rawKeys) == 0 {
		return iter(ast.NullTerm())
	}

	key, err := jwk.New(rawKeys[0])
	if err != nil {
		return err
	}

	jsonKey, err := json.Marshal(key)
	if err != nil {
		return err
	}

	if err := util.UnmarshalJSON(jsonKey, &x); err != nil {
		return err
	}

	value, err := ast.InterfaceToValue(x)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(value))
}

func builtinCryptoParsePrivateKeys(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	a := operands[0].Value
	input, err := builtins.StringOperand(a, 1)
	if err != nil {
		return err
	}

	if string(input) == "" {
		return iter(ast.NullTerm())
	}

	// get the raw private key
	rawKeys, err := getPrivateKeysFromPEMData(string(input))
	if err != nil {
		return err
	}

	if len(rawKeys) == 0 {
		return iter(ast.NewTerm(ast.NewArray()))
	}

	bs, err := json.Marshal(rawKeys)
	if err != nil {
		return err
	}

	var x interface{}
	if err := util.UnmarshalJSON(bs, &x); err != nil {
		return err
	}

	value, err := ast.InterfaceToValue(x)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(value))
}

func hashHelper(a ast.Value, h func(ast.String) string) (ast.Value, error) {
	s, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	return ast.String(h(s)), nil
}

func builtinCryptoMd5(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	res, err := hashHelper(operands[0].Value, func(s ast.String) string { return fmt.Sprintf("%x", md5.Sum([]byte(s))) })
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(res))
}

func builtinCryptoSha1(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	res, err := hashHelper(operands[0].Value, func(s ast.String) string { return fmt.Sprintf("%x", sha1.Sum([]byte(s))) })
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(res))
}

func builtinCryptoSha256(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	res, err := hashHelper(operands[0].Value, func(s ast.String) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) })
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(res))
}

func hmacHelper(operands []*ast.Term, iter func(*ast.Term) error, h func() hash.Hash) error {
	a1 := operands[0].Value
	message, err := builtins.StringOperand(a1, 1)
	if err != nil {
		return err
	}

	a2 := operands[1].Value
	key, err := builtins.StringOperand(a2, 2)
	if err != nil {
		return err
	}

	mac := hmac.New(h, []byte(key))
	mac.Write([]byte(message))
	messageDigest := mac.Sum(nil)

	return iter(ast.StringTerm(fmt.Sprintf("%x", messageDigest)))
}

func builtinCryptoHmacMd5(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	return hmacHelper(operands, iter, md5.New)
}

func builtinCryptoHmacSha1(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	return hmacHelper(operands, iter, sha1.New)
}

func builtinCryptoHmacSha256(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	return hmacHelper(operands, iter, sha256.New)
}

func builtinCryptoHmacSha512(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	return hmacHelper(operands, iter, sha512.New)
}

func builtinCryptoHmacEqual(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	a1 := operands[0].Value
	mac1, err := builtins.StringOperand(a1, 1)
	if err != nil {
		return err
	}

	a2 := operands[1].Value
	mac2, err := builtins.StringOperand(a2, 2)
	if err != nil {
		return err
	}

	res := hmac.Equal([]byte(mac1), []byte(mac2))

	return iter(ast.BooleanTerm(res))
}

func init() {
	RegisterBuiltinFunc(ast.CryptoX509ParseCertificates.Name, builtinCryptoX509ParseCertificates)
	RegisterBuiltinFunc(ast.CryptoX509ParseAndVerifyCertificates.Name, builtinCryptoX509ParseAndVerifyCertificates)
	RegisterBuiltinFunc(ast.CryptoX509ParseAndVerifyCertificatesWithOptions.Name, builtinCryptoX509ParseAndVerifyCertificatesWithOptions)
	RegisterBuiltinFunc(ast.CryptoMd5.Name, builtinCryptoMd5)
	RegisterBuiltinFunc(ast.CryptoSha1.Name, builtinCryptoSha1)
	RegisterBuiltinFunc(ast.CryptoSha256.Name, builtinCryptoSha256)
	RegisterBuiltinFunc(ast.CryptoX509ParseCertificateRequest.Name, builtinCryptoX509ParseCertificateRequest)
	RegisterBuiltinFunc(ast.CryptoX509ParseRSAPrivateKey.Name, builtinCryptoJWKFromPrivateKey)
	RegisterBuiltinFunc(ast.CryptoParsePrivateKeys.Name, builtinCryptoParsePrivateKeys)
	RegisterBuiltinFunc(ast.CryptoX509ParseKeyPair.Name, builtinCryptoX509ParseKeyPair)
	RegisterBuiltinFunc(ast.CryptoHmacMd5.Name, builtinCryptoHmacMd5)
	RegisterBuiltinFunc(ast.CryptoHmacSha1.Name, builtinCryptoHmacSha1)
	RegisterBuiltinFunc(ast.CryptoHmacSha256.Name, builtinCryptoHmacSha256)
	RegisterBuiltinFunc(ast.CryptoHmacSha512.Name, builtinCryptoHmacSha512)
	RegisterBuiltinFunc(ast.CryptoHmacEqual.Name, builtinCryptoHmacEqual)
}

func verifyX509CertificateChain(certs []*x509.Certificate, vo x509.VerifyOptions) ([]*x509.Certificate, error) {
	if len(certs) < 2 {
		return nil, builtins.NewOperandErr(1, "must supply at least two certificates to be able to verify")
	}

	// first cert is the root
	roots := x509.NewCertPool()
	roots.AddCert(certs[0])

	// all other certs except the last are intermediates
	intermediates := x509.NewCertPool()
	for i := 1; i < len(certs)-1; i++ {
		intermediates.AddCert(certs[i])
	}

	// last cert is the leaf
	leaf := certs[len(certs)-1]

	// verify the cert chain back to the root
	verifyOpts := x509.VerifyOptions{
		Roots:                     roots,
		Intermediates:             intermediates,
		DNSName:                   vo.DNSName,
		CurrentTime:               vo.CurrentTime,
		KeyUsages:                 vo.KeyUsages,
		MaxConstraintComparisions: vo.MaxConstraintComparisions,
	}
	chains, err := leaf.Verify(verifyOpts)
	if err != nil {
		return nil, err
	}

	return chains[0], nil
}

func getX509CertsFromString(certs string) ([]*x509.Certificate, error) {
	// if the input is PEM handle that
	if strings.HasPrefix(certs, "-----BEGIN") {
		return getX509CertsFromPem([]byte(certs))
	}

	// assume input is base64 if not PEM
	b64, err := base64.StdEncoding.DecodeString(certs)
	if err != nil {
		return nil, err
	}

	// handle if the decoded base64 contains PEM rather than the expected DER
	if bytes.HasPrefix(b64, []byte("-----BEGIN")) {
		return getX509CertsFromPem(b64)
	}

	// otherwise assume the contents are DER
	return x509.ParseCertificates(b64)
}

func getX509CertsFromPem(pemBlocks []byte) ([]*x509.Certificate, error) {
	var decodedCerts []byte
	for len(pemBlocks) > 0 {
		p, r := pem.Decode(pemBlocks)
		if p != nil && p.Type != blockTypeCertificate {
			return nil, fmt.Errorf("PEM block type is '%s', expected %s", p.Type, blockTypeCertificate)
		}

		if p == nil {
			break
		}

		pemBlocks = r
		decodedCerts = append(decodedCerts, p.Bytes...)
	}

	return x509.ParseCertificates(decodedCerts)
}

func getPrivateKeysFromPEMData(pemData string) ([]crypto.PrivateKey, error) {
	pemBlockString := pemData

	var validPrivateKeys []crypto.PrivateKey

	// if the input is base64, decode it
	bs, err := base64.StdEncoding.DecodeString(pemBlockString)
	if err == nil {
		pemBlockString = string(bs)
	}
	bs = []byte(pemBlockString)

	for len(bs) > 0 {
		inputLen := len(bs)
		var block *pem.Block
		block, bs = pem.Decode(bs)
		if block == nil && len(bs) == 0 {
			break
		}
		// should only happen if end of input is not a valid PEM block. See TestParseRSAPrivateKeyVariedPemInput.
		if inputLen == len(bs) {
			break
		}

		if block == nil {
			continue
		}

		switch block.Type {
		case blockTypeRSAPrivateKey:
			parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			validPrivateKeys = append(validPrivateKeys, parsedKey)
		case blockTypePrivateKey:
			parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			validPrivateKeys = append(validPrivateKeys, parsedKey)
		case blockTypeEcPrivateKey:
			parsedKey, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			validPrivateKeys = append(validPrivateKeys, parsedKey)
		}
	}
	return validPrivateKeys, nil
}

// addCACertsFromFile adds CA certificates from filePath into the given pool.
// If pool is nil, it creates a new x509.CertPool. pool is returned.
func addCACertsFromFile(pool *x509.CertPool, filePath string) (*x509.CertPool, error) {
	if pool == nil {
		pool = x509.NewCertPool()
	}

	caCert, err := readCertFromFile(filePath)
	if err != nil {
		return nil, err
	}

	if ok := pool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("could not append CA certificates from %q", filePath)
	}

	return pool, nil
}

// addCACertsFromBytes adds CA certificates from pemBytes into the given pool.
// If pool is nil, it creates a new x509.CertPool. pool is returned.
func addCACertsFromBytes(pool *x509.CertPool, pemBytes []byte) (*x509.CertPool, error) {
	if pool == nil {
		pool = x509.NewCertPool()
	}

	if ok := pool.AppendCertsFromPEM(pemBytes); !ok {
		return nil, fmt.Errorf("could not append certificates")
	}

	return pool, nil
}

// addCACertsFromEnv adds CA certificates from the environment variable named
// by envName into the given pool. If pool is nil, it creates a new x509.CertPool.
// pool is returned.
func addCACertsFromEnv(pool *x509.CertPool, envName string) (*x509.CertPool, error) {
	pool, err := addCACertsFromBytes(pool, []byte(os.Getenv(envName)))
	if err != nil {
		return nil, fmt.Errorf("could not add CA certificates from envvar %q: %w", envName, err)
	}

	return pool, err
}

// ReadCertFromFile reads a cert from file
func readCertFromFile(localCertFile string) ([]byte, error) {
	// Read in the cert file
	certPEM, err := os.ReadFile(localCertFile)
	if err != nil {
		return nil, err
	}
	return certPEM, nil
}

func getTLSx509KeyPairFromString(certPemBlock []byte, keyPemBlock []byte) (*tls.Certificate, error) {

	if !strings.HasPrefix(string(certPemBlock), "-----BEGIN") {
		s, err := base64.StdEncoding.DecodeString(string(certPemBlock))
		if err != nil {
			return nil, err
		}
		certPemBlock = s
	}

	if !strings.HasPrefix(string(keyPemBlock), "-----BEGIN") {
		s, err := base64.StdEncoding.DecodeString(string(keyPemBlock))
		if err != nil {
			return nil, err
		}
		keyPemBlock = s
	}

	// we assume it a DER certificate and try to convert it to a PEM.
	if !bytes.HasPrefix(certPemBlock, []byte("-----BEGIN")) {

		pemBlock := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certPemBlock,
		}

		var buf bytes.Buffer
		if err := pem.Encode(&buf, pemBlock); err != nil {
			return nil, err
		}
		certPemBlock = buf.Bytes()

	}
	// we assume it a DER key and try to convert it to a PEM.
	if !bytes.HasPrefix(keyPemBlock, []byte("-----BEGIN")) {
		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyPemBlock,
		}
		var buf bytes.Buffer
		if err := pem.Encode(&buf, pemBlock); err != nil {
			return nil, err
		}
		keyPemBlock = buf.Bytes()
	}

	cert, err := tls.X509KeyPair(certPemBlock, keyPemBlock)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// ReadKeyFromFile reads a key from file
func readKeyFromFile(localKeyFile string) ([]byte, error) {
	// Read in the cert file
	key, err := os.ReadFile(localKeyFile)
	if err != nil {
		return nil, err
	}
	return key, nil
}
