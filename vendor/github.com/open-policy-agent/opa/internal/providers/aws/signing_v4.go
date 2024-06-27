// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package aws

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
)

func stringFromTerm(t *ast.Term) string {
	if v, ok := t.Value.(ast.String); ok {
		return string(v)
	}
	return ""
}

// Headers that may be mutated before reaching an aws service (eg by a proxy) should be added here to omit them from
// the sigv4 canonical request
// ref. https://github.com/aws/aws-sdk-go/blob/master/aws/signer/v4/v4.go#L92
var awsSigv4IgnoredHeaders = map[string]struct{}{
	"authorization":   {},
	"user-agent":      {},
	"x-amzn-trace-id": {},
}

type Credentials struct {
	AccessKey    string
	SecretKey    string
	RegionName   string
	SessionToken string
}

func CredentialsFromObject(v ast.Object) Credentials {
	var creds Credentials
	awsAccessKey := v.Get(ast.StringTerm("aws_access_key"))
	awsSecretKey := v.Get(ast.StringTerm("aws_secret_access_key"))
	awsRegion := v.Get(ast.StringTerm("aws_region"))
	awsSessionToken := v.Get(ast.StringTerm("aws_session_token"))

	creds.AccessKey = stringFromTerm(awsAccessKey)
	creds.SecretKey = stringFromTerm(awsSecretKey)
	creds.RegionName = stringFromTerm(awsRegion)
	if awsSessionToken != nil {
		creds.SessionToken = stringFromTerm(awsSessionToken)
	}
	return creds
}

func sha256MAC(message string, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func sortKeys(strMap map[string][]string) []string {
	keys := make([]string, len(strMap))

	i := 0
	for k := range strMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

// SignRequest modifies an http.Request to include an AWS V4 signature based on the provided credentials.
func SignRequest(req *http.Request, service string, creds Credentials, theTime time.Time, sigVersion string) error {
	// General ref. https://docs.aws.amazon.com/general/latest/gr/sigv4_signing.html
	// S3 ref. https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
	// APIGateway ref. https://docs.aws.amazon.com/apigateway/api-reference/signing-requests/

	var body []byte
	if req.Body == nil {
		body = []byte("")
	} else {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return errors.New("error getting request body: " + err.Error())
		}
		// Since ReadAll consumed the body ReadCloser, we must create a new ReadCloser for the request so that the
		// subsequent read starts from the beginning
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	now := theTime.UTC()

	if sigVersion == "4a" {
		signedHeaders := SignV4a(req.Header, req.Method, req.URL, body, service, creds, now)
		req.Header = signedHeaders
	} else {
		authHeader, awsHeaders := SignV4(req.Header, req.Method, req.URL, body, service, creds, now)
		req.Header.Set("Authorization", authHeader)
		for k, v := range awsHeaders {
			req.Header.Add(k, v)
		}
	}

	return nil
}

// SignV4 modifies a map[string][]string of headers to generate an AWS V4 signature + headers based on the config/credentials provided.
func SignV4(headers map[string][]string, method string, theURL *url.URL, body []byte, service string, awsCreds Credentials, theTime time.Time) (string, map[string]string) {
	// General ref. https://docs.aws.amazon.com/general/latest/gr/sigv4_signing.html
	// S3 ref. https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
	// APIGateway ref. https://docs.aws.amazon.com/apigateway/api-reference/signing-requests/
	bodyHexHash := fmt.Sprintf("%x", sha256.Sum256(body))

	now := theTime.UTC()

	// V4 signing has specific ideas of how it wants to see dates/times encoded
	dateNow := now.Format("20060102")
	iso8601Now := now.Format("20060102T150405Z")

	awsHeaders := map[string]string{
		"host":       theURL.Host,
		"x-amz-date": iso8601Now,
	}

	// s3 and glacier require the extra x-amz-content-sha256 header. other services do not.
	if service == "s3" || service == "glacier" {
		awsHeaders["x-amz-content-sha256"] = bodyHexHash
	}

	// the security token header is necessary for ephemeral credentials, e.g. from
	// the EC2 metadata service
	if awsCreds.SessionToken != "" {
		awsHeaders["x-amz-security-token"] = awsCreds.SessionToken
	}

	headersToSign := map[string][]string{}
	// sign all of the aws headers.
	for k, v := range awsHeaders {
		headersToSign[k] = []string{v}
	}

	// sign all of the request's headers, except for those in the ignore list
	for k, v := range headers {
		lowercaseHeader := strings.ToLower(k)
		if _, ok := awsSigv4IgnoredHeaders[lowercaseHeader]; !ok {
			headersToSign[lowercaseHeader] = v
		}
	}

	// the "canonical request" is the normalized version of the AWS service access
	// that we're attempting to perform
	canonicalReq := method + "\n"               // HTTP method
	canonicalReq += theURL.EscapedPath() + "\n" // URI-escaped path
	canonicalReq += theURL.RawQuery + "\n"      // RAW Query String

	// include the values for the signed headers
	orderedKeys := sortKeys(headersToSign)
	for _, k := range orderedKeys {
		canonicalReq += k + ":" + strings.Join(headersToSign[k], ",") + "\n"
	}
	canonicalReq += "\n" // linefeed to terminate headers

	// include the list of the signed headers
	headerList := strings.Join(orderedKeys, ";")
	canonicalReq += headerList + "\n"
	canonicalReq += bodyHexHash

	// the "string to sign" is a time-bounded, scoped request token which
	// is linked to the "canonical request" by inclusion of its SHA-256 hash
	strToSign := "AWS4-HMAC-SHA256\n"                                                    // V4 signing with SHA-256 HMAC
	strToSign += iso8601Now + "\n"                                                       // ISO 8601 time
	strToSign += dateNow + "/" + awsCreds.RegionName + "/" + service + "/aws4_request\n" // scoping for signature
	strToSign += fmt.Sprintf("%x", sha256.Sum256([]byte(canonicalReq)))                  // SHA-256 of canonical request

	// the "signing key" is generated by repeated HMAC-SHA256 based on the same
	// scoping that's included in the "string to sign"; but including the secret key
	// to allow AWS to validate it
	signingKey := sha256MAC(dateNow, []byte("AWS4"+awsCreds.SecretKey))
	signingKey = sha256MAC(awsCreds.RegionName, signingKey)
	signingKey = sha256MAC(service, signingKey)
	signingKey = sha256MAC("aws4_request", signingKey)

	// the "signature" is finally the "string to sign" signed by the "signing key"
	signature := sha256MAC(strToSign, signingKey)

	// required format of Authorization header; n.b. the access key corresponding to
	// the secret key is included here
	authHeader := "AWS4-HMAC-SHA256 Credential=" + awsCreds.AccessKey + "/" + dateNow
	authHeader += "/" + awsCreds.RegionName + "/" + service + "/aws4_request,"
	authHeader += "SignedHeaders=" + headerList + ","
	authHeader += "Signature=" + fmt.Sprintf("%x", signature)

	return authHeader, awsHeaders
}
