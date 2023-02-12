// modified from github.com/aws/aws-sdk-go-v2/internal/v4a@7a32d707af
package aws

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	signerCrypto "github.com/open-policy-agent/opa/internal/providers/aws/crypto"
	v4Internal "github.com/open-policy-agent/opa/internal/providers/aws/v4"
)

const (
	// AmzRegionSetKey represents the region set header used for sigv4a
	AmzRegionSetKey     = "X-Amz-Region-Set"
	amzSecurityTokenKey = v4Internal.AmzSecurityTokenKey
	amzDateKey          = v4Internal.AmzDateKey
	authorizationHeader = "Authorization"

	signingAlgorithm = "AWS4-ECDSA-P256-SHA256"

	timeFormat      = "20060102T150405Z"
	shortTimeFormat = "20060102"
)

var (
	p256          elliptic.Curve
	nMinusTwoP256 *big.Int

	one = new(big.Int).SetInt64(1)

	cache = credsCache{}

	randomSource = rand.Reader
)

func init() {
	// Ensure the elliptic curve parameters are initialized on package import rather then on first usage
	p256 = elliptic.P256()

	nMinusTwoP256 = new(big.Int).SetBytes(p256.Params().N.Bytes())
	nMinusTwoP256 = nMinusTwoP256.Sub(nMinusTwoP256, new(big.Int).SetInt64(2))
}

type credsCache struct {
	asymmetric atomic.Value
	m          sync.Mutex
}

// SetRandomSource used for testing to override rand so tests can expect stable output
func SetRandomSource(reader io.Reader) {
	randomSource = reader
}

// deriveKeyFromAccessKeyPair derives a NIST P-256 PrivateKey from the given
// IAM AccessKey and SecretKey pair.
//
// Based on FIPS.186-4 Appendix B.4.2
func deriveKeyFromAccessKeyPair(accessKey, secretKey string) (*ecdsa.PrivateKey, error) {
	params := p256.Params()
	bitLen := params.BitSize // Testing random candidates does not require an additional 64 bits
	counter := 0x01

	buffer := make([]byte, 1+len(accessKey)) // 1 byte counter + len(accessKey)
	kdfContext := bytes.NewBuffer(buffer)

	inputKey := append([]byte("AWS4A"), []byte(secretKey)...)

	d := new(big.Int)
	for {
		kdfContext.Reset()
		kdfContext.WriteString(accessKey)
		kdfContext.WriteByte(byte(counter))

		key, err := signerCrypto.HMACKeyDerivation(sha256.New, bitLen, inputKey, []byte(signingAlgorithm), kdfContext.Bytes())
		if err != nil {
			return nil, err
		}

		// Check key first before calling SetBytes if key is in fact a valid candidate.
		// This ensures the byte slice is the correct length (32-bytes) to compare in constant-time
		cmp, err := signerCrypto.ConstantTimeByteCompare(key, nMinusTwoP256.Bytes())
		if err != nil {
			return nil, err
		}
		if cmp == -1 {
			d.SetBytes(key)
			break
		}

		counter++
		if counter > 0xFF {
			return nil, fmt.Errorf("exhausted single byte external counter")
		}
	}
	d = d.Add(d, one)

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = p256
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = p256.ScalarBaseMult(d.Bytes())

	return priv, nil
}

// v4aCredentials is Context, ECDSA, and Optional Session Token that can be used
// to sign requests using SigV4a
type v4aCredentials struct {
	Context      string
	PrivateKey   *ecdsa.PrivateKey
	SessionToken string
}

// retrievePrivateKey returns credentials suitable for SigV4a signing
func retrievePrivateKey(symmetric Credentials) (v4aCredentials, error) {
	cache.m.Lock()
	defer cache.m.Unlock()

	// try to get creds from cache
	v := cache.asymmetric.Load()
	if v != nil {
		c := v.(*v4aCredentials)
		// if the cached Context matches the symmetric AccessKey ID, then use cached value. Otherwise, creds have
		// changed and we need to derive new asymmetric creds
		if c != nil && c.Context == symmetric.AccessKey {
			return *c, nil
		}
	}

	privateKey, err := deriveKeyFromAccessKeyPair(symmetric.AccessKey, symmetric.SecretKey)
	if err != nil {
		return v4aCredentials{}, fmt.Errorf("failed to derive asymmetric key from credentials")
	}

	creds := v4aCredentials{
		Context:      symmetric.AccessKey,
		PrivateKey:   privateKey,
		SessionToken: symmetric.SessionToken,
	}

	// cache derived asymmetric creds so we don't derive new ones until symmetric creds change
	cache.asymmetric.Store(&creds)

	return creds, nil
}

type httpSigner struct {
	Request     *http.Request
	ServiceName string
	RegionSet   []string
	Time        time.Time
	Credentials v4aCredentials

	// PayloadHash is the hex encoded SHA-256 hash of the request payload
	// If len(PayloadHash) == 0 the signer will attempt to send the request
	// as an unsigned payload. Note: Unsigned payloads only work for a subset of services.
	PayloadHash string
}

func (s *httpSigner) setRequiredSigningFields(headers http.Header, query url.Values) {
	amzDate := s.Time.Format(timeFormat)

	headers.Set(AmzRegionSetKey, strings.Join(s.RegionSet, ","))
	headers.Set(amzDateKey, amzDate)
	if len(s.Credentials.SessionToken) > 0 {
		headers.Set(amzSecurityTokenKey, s.Credentials.SessionToken)
	}
}

// Build modifies the Request attribute of the httpSigner, adding an Authorization header
func (s *httpSigner) Build() (signedRequest, error) {
	req := s.Request

	query := req.URL.Query()
	headers := req.Header

	// seemingly required by S3/MRAP -- 403 Forbidden otherwise
	headers.Set("host", req.URL.Host)
	headers.Set("x-amz-content-sha256", s.PayloadHash)

	s.setRequiredSigningFields(headers, query)

	// Sort Each Query Key's Values
	for key := range query {
		sort.Strings(query[key])
	}

	v4Internal.SanitizeHostForHeader(req)

	credentialScope := s.buildCredentialScope()
	credentialStr := s.Credentials.Context + "/" + credentialScope

	unsignedHeaders := headers

	host := req.URL.Host
	if len(req.Host) > 0 {
		host = req.Host
	}

	signedHeaders, signedHeadersStr, canonicalHeaderStr := s.buildCanonicalHeaders(host, v4Internal.IgnoredHeaders, unsignedHeaders, s.Request.ContentLength)

	rawQuery := strings.Replace(query.Encode(), "+", "%20", -1)

	canonicalURI := v4Internal.GetURIPath(req.URL)

	canonicalString := s.buildCanonicalString(
		req.Method,
		canonicalURI,
		rawQuery,
		signedHeadersStr,
		canonicalHeaderStr,
	)

	strToSign := s.buildStringToSign(credentialScope, canonicalString)
	signingSignature, err := s.buildSignature(strToSign)
	if err != nil {
		return signedRequest{}, err
	}

	headers[authorizationHeader] = append(headers[authorizationHeader][:0], buildAuthorizationHeader(credentialStr, signedHeadersStr, signingSignature))

	req.URL.RawQuery = rawQuery

	return signedRequest{
		Request:         req,
		SignedHeaders:   signedHeaders,
		CanonicalString: canonicalString,
		StringToSign:    strToSign,
	}, nil
}

func (s *httpSigner) buildCredentialScope() string {
	return strings.Join([]string{
		s.Time.Format(shortTimeFormat),
		s.ServiceName,
		"aws4_request",
	}, "/")

}

func buildAuthorizationHeader(credentialStr, signedHeadersStr, signingSignature string) string {
	const credential = "Credential="
	const signedHeaders = "SignedHeaders="
	const signature = "Signature="
	const commaSpace = ", "

	var parts strings.Builder
	parts.Grow(len(signingAlgorithm) + 1 +
		len(credential) + len(credentialStr) + len(commaSpace) +
		len(signedHeaders) + len(signedHeadersStr) + len(commaSpace) +
		len(signature) + len(signingSignature),
	)
	parts.WriteString(signingAlgorithm)
	parts.WriteRune(' ')
	parts.WriteString(credential)
	parts.WriteString(credentialStr)
	parts.WriteString(commaSpace)
	parts.WriteString(signedHeaders)
	parts.WriteString(signedHeadersStr)
	parts.WriteString(commaSpace)
	parts.WriteString(signature)
	parts.WriteString(signingSignature)
	return parts.String()
}

func (s *httpSigner) buildCanonicalHeaders(host string, rule v4Internal.Rule, header http.Header, length int64) (signed http.Header, signedHeaders, canonicalHeadersStr string) {
	signed = make(http.Header)

	const hostHeader = "host"
	headers := make([]string, 0)

	if length > 0 {
		const contentLengthHeader = "content-length"
		headers = append(headers, contentLengthHeader)
		signed[contentLengthHeader] = append(signed[contentLengthHeader], strconv.FormatInt(length, 10))
	}

	for k, v := range header {
		if !rule.IsValid(k) {
			continue // ignored header
		}

		lowerCaseKey := strings.ToLower(k)
		if _, ok := signed[lowerCaseKey]; ok {
			// include additional values
			signed[lowerCaseKey] = append(signed[lowerCaseKey], v...)
			continue
		}

		headers = append(headers, lowerCaseKey)
		signed[lowerCaseKey] = v
	}
	sort.Strings(headers)

	signedHeaders = strings.Join(headers, ";")

	var canonicalHeaders strings.Builder
	n := len(headers)
	const colon = ':'
	for i := 0; i < n; i++ {
		if headers[i] == hostHeader {
			canonicalHeaders.WriteString(hostHeader)
			canonicalHeaders.WriteRune(colon)
			canonicalHeaders.WriteString(v4Internal.StripExcessSpaces(host))
		} else {
			canonicalHeaders.WriteString(headers[i])
			canonicalHeaders.WriteRune(colon)
			// Trim out leading, trailing, and dedup inner spaces from signed header values.
			values := signed[headers[i]]
			for j, v := range values {
				cleanedValue := strings.TrimSpace(v4Internal.StripExcessSpaces(v))
				canonicalHeaders.WriteString(cleanedValue)
				if j < len(values)-1 {
					canonicalHeaders.WriteRune(',')
				}
			}
		}
		canonicalHeaders.WriteRune('\n')
	}
	canonicalHeadersStr = canonicalHeaders.String()

	return signed, signedHeaders, canonicalHeadersStr
}

func (s *httpSigner) buildCanonicalString(method, uri, query, signedHeaders, canonicalHeaders string) string {
	return strings.Join([]string{
		method,
		uri,
		query,
		canonicalHeaders,
		signedHeaders,
		s.PayloadHash,
	}, "\n")
}

func (s *httpSigner) buildStringToSign(credentialScope, canonicalRequestString string) string {
	return strings.Join([]string{
		signingAlgorithm,
		s.Time.Format(timeFormat),
		credentialScope,
		hex.EncodeToString(makeHash(sha256.New(), []byte(canonicalRequestString))),
	}, "\n")
}

func makeHash(hash hash.Hash, b []byte) []byte {
	hash.Reset()
	hash.Write(b)
	return hash.Sum(nil)
}

func (s *httpSigner) buildSignature(strToSign string) (string, error) {
	sig, err := s.Credentials.PrivateKey.Sign(randomSource, makeHash(sha256.New(), []byte(strToSign)), crypto.SHA256)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

type signedRequest struct {
	Request         *http.Request
	SignedHeaders   http.Header
	CanonicalString string
	StringToSign    string
}

// SignV4a returns a map[string][]string of headers, including an added AWS V4a signature based on the config/credentials provided.
func SignV4a(headers map[string][]string, method string, theURL *url.URL, body []byte, service string, awsCreds Credentials, theTime time.Time) map[string][]string {
	bodyHexHash := fmt.Sprintf("%x", sha256.Sum256(body))

	key, err := retrievePrivateKey(awsCreds)
	if err != nil {
		return map[string][]string{}
	}

	bodyReader := bytes.NewReader(body)
	req, _ := http.NewRequest(method, theURL.String(), bodyReader)
	req.Header = headers

	signer := &httpSigner{
		Request:     req,
		PayloadHash: bodyHexHash,
		ServiceName: service,
		RegionSet:   []string{"*"},
		Credentials: key,
		Time:        theTime,
	}

	_, err = signer.Build()
	if err != nil {
		return map[string][]string{}
	}

	return req.Header
}
