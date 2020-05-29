// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/open-policy-agent/opa/internal/version"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

const defaultHTTPRequestTimeoutEnv = "HTTP_SEND_TIMEOUT"

var defaultHTTPRequestTimeout = time.Second * 5

var allowedKeyNames = [...]string{
	"method",
	"url",
	"body",
	"enable_redirect",
	"force_json_decode",
	"headers",
	"raw_body",
	"tls_use_system_certs",
	"tls_ca_cert_file",
	"tls_ca_cert_env_variable",
	"tls_client_cert_env_variable",
	"tls_client_key_env_variable",
	"tls_client_cert_file",
	"tls_client_key_file",
	"tls_insecure_skip_verify",
	"timeout",
}
var allowedKeys = ast.NewSet()

var requiredKeys = ast.NewSet(ast.StringTerm("method"), ast.StringTerm("url"))

type httpSendKey string

// httpSendBuiltinCacheKey is the key in the builtin context cache that
// points to the http.send() specific cache resides at.
const httpSendBuiltinCacheKey httpSendKey = "HTTP_SEND_CACHE_KEY"

func builtinHTTPSend(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {

	req, err := validateHTTPRequestOperand(args[0], 1)
	if err != nil {
		return handleBuiltinErr(ast.HTTPSend.Name, bctx.Location, err)
	}

	// check if cache already has a response for this query
	resp := checkHTTPSendCache(bctx, req)
	if resp == nil {
		var err error
		resp, err = executeHTTPRequest(bctx, req)
		if err != nil {
			return handleHTTPSendErr(bctx, err)
		}

		// add result to cache
		insertIntoHTTPSendCache(bctx, req, resp)
	}

	return iter(ast.NewTerm(resp))
}

func init() {
	createAllowedKeys()
	initDefaults()
	RegisterBuiltinFunc(ast.HTTPSend.Name, builtinHTTPSend)
}

func handleHTTPSendErr(bctx BuiltinContext, err error) error {
	// Return HTTP client timeout errors in a generic error message to avoid confusion about what happened.
	// Do not do this if the builtin context was cancelled and is what caused the request to stop.
	if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() && bctx.Context.Err() == nil {
		err = fmt.Errorf("%s %s: request timed out", urlErr.Op, urlErr.URL)
	}
	return handleBuiltinErr(ast.HTTPSend.Name, bctx.Location, err)
}

func initDefaults() {
	timeoutDuration := os.Getenv(defaultHTTPRequestTimeoutEnv)
	if timeoutDuration != "" {
		var err error
		defaultHTTPRequestTimeout, err = time.ParseDuration(timeoutDuration)
		if err != nil {
			// If it is set to something not valid don't let the process continue in a state
			// that will almost definitely give unexpected results by having it set at 0
			// which means no timeout..
			// This environment variable isn't considered part of the public API.
			// TODO(patrick-east): Remove the environment variable
			panic(fmt.Sprintf("invalid value for HTTP_SEND_TIMEOUT: %s", err))
		}
	}
}

func validateHTTPRequestOperand(term *ast.Term, pos int) (ast.Object, error) {

	obj, err := builtins.ObjectOperand(term.Value, pos)
	if err != nil {
		return nil, err
	}

	requestKeys := ast.NewSet(obj.Keys()...)

	invalidKeys := requestKeys.Diff(allowedKeys)
	if invalidKeys.Len() != 0 {
		return nil, builtins.NewOperandErr(pos, "invalid request parameters(s): %v", invalidKeys)
	}

	missingKeys := requiredKeys.Diff(requestKeys)
	if missingKeys.Len() != 0 {
		return nil, builtins.NewOperandErr(pos, "missing required request parameters(s): %v", missingKeys)
	}

	return obj, nil

}

// Adds custom headers to a new HTTP request.
func addHeaders(req *http.Request, headers map[string]interface{}) (bool, error) {
	for k, v := range headers {
		// Type assertion
		header, ok := v.(string)
		if !ok {
			return false, fmt.Errorf("invalid type for headers value %q", v)
		}

		// If the Host header is given, bump that up to
		// the request. Otherwise, just collect it in the
		// headers.
		k := http.CanonicalHeaderKey(k)
		switch k {
		case "Host":
			req.Host = header
		default:
			req.Header.Add(k, header)
		}
	}

	return true, nil
}

func executeHTTPRequest(bctx BuiltinContext, obj ast.Object) (ast.Value, error) {
	var url string
	var method string
	var tlsCaCertEnvVar []byte
	var tlsCaCertFile string
	var tlsClientKeyEnvVar []byte
	var tlsClientCertEnvVar []byte
	var tlsClientCertFile string
	var tlsClientKeyFile string
	var body *bytes.Buffer
	var rawBody *bytes.Buffer
	var enableRedirect bool
	var forceJSONDecode bool
	var tlsUseSystemCerts bool
	var tlsConfig tls.Config
	var clientCerts []tls.Certificate
	var customHeaders map[string]interface{}
	var tlsInsecureSkipVerify bool
	var timeout = defaultHTTPRequestTimeout

	for _, val := range obj.Keys() {
		key, err := ast.JSON(val.Value)
		if err != nil {
			return nil, err
		}
		key = key.(string)

		switch key {
		case "method":
			method = obj.Get(val).String()
			method = strings.ToUpper(strings.Trim(method, "\""))
		case "url":
			url = obj.Get(val).String()
			url = strings.Trim(url, "\"")
		case "enable_redirect":
			enableRedirect, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, err
			}
		case "force_json_decode":
			forceJSONDecode, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, err
			}
		case "body":
			bodyVal := obj.Get(val).Value
			bodyValInterface, err := ast.JSON(bodyVal)
			if err != nil {
				return nil, err
			}

			bodyValBytes, err := json.Marshal(bodyValInterface)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(bodyValBytes)
		case "raw_body":
			s, ok := obj.Get(val).Value.(ast.String)
			if !ok {
				return nil, fmt.Errorf("raw_body must be a string")
			}
			rawBody = bytes.NewBuffer([]byte(s))
		case "tls_use_system_certs":
			tlsUseSystemCerts, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, err
			}
		case "tls_ca_cert_file":
			tlsCaCertFile = obj.Get(val).String()
			tlsCaCertFile = strings.Trim(tlsCaCertFile, "\"")
		case "tls_ca_cert_env_variable":
			caCertEnv := obj.Get(val).String()
			caCertEnv = strings.Trim(caCertEnv, "\"")
			tlsCaCertEnvVar = []byte(os.Getenv(caCertEnv))
		case "tls_client_cert_env_variable":
			clientCertEnv := obj.Get(val).String()
			clientCertEnv = strings.Trim(clientCertEnv, "\"")
			tlsClientCertEnvVar = []byte(os.Getenv(clientCertEnv))
		case "tls_client_key_env_variable":
			clientKeyEnv := obj.Get(val).String()
			clientKeyEnv = strings.Trim(clientKeyEnv, "\"")
			tlsClientKeyEnvVar = []byte(os.Getenv(clientKeyEnv))
		case "tls_client_cert_file":
			tlsClientCertFile = obj.Get(val).String()
			tlsClientCertFile = strings.Trim(tlsClientCertFile, "\"")
		case "tls_client_key_file":
			tlsClientKeyFile = obj.Get(val).String()
			tlsClientKeyFile = strings.Trim(tlsClientKeyFile, "\"")
		case "headers":
			headersVal := obj.Get(val).Value
			headersValInterface, err := ast.JSON(headersVal)
			if err != nil {
				return nil, err
			}
			var ok bool
			customHeaders, ok = headersValInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid type for headers key")
			}
		case "tls_insecure_skip_verify":
			tlsInsecureSkipVerify, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, err
			}
		case "timeout":
			timeout, err = parseTimeout(obj.Get(val).Value)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid parameter %q", key)
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	if tlsInsecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsInsecureSkipVerify},
		}
	}
	if tlsClientCertFile != "" && tlsClientKeyFile != "" {
		clientCertFromFile, err := tls.LoadX509KeyPair(tlsClientCertFile, tlsClientKeyFile)
		if err != nil {
			return nil, err
		}
		clientCerts = append(clientCerts, clientCertFromFile)
	}

	if len(tlsClientCertEnvVar) > 0 && len(tlsClientKeyEnvVar) > 0 {
		clientCertFromEnv, err := tls.X509KeyPair(tlsClientCertEnvVar, tlsClientKeyEnvVar)
		if err != nil {
			return nil, err
		}
		clientCerts = append(clientCerts, clientCertFromEnv)
	}

	isTLS := false
	if len(clientCerts) > 0 {
		isTLS = true
		tlsConfig.Certificates = append(tlsConfig.Certificates, clientCerts...)
	}

	if tlsUseSystemCerts || len(tlsCaCertFile) > 0 || len(tlsCaCertEnvVar) > 0 {
		isTLS = true
		connRootCAs, err := createRootCAs(tlsCaCertFile, tlsCaCertEnvVar, tlsUseSystemCerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = connRootCAs
	}

	if isTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tlsConfig,
		}
	}

	// check if redirects are enabled
	if !enableRedirect {
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if rawBody != nil {
		body = rawBody
	} else if body == nil {
		body = bytes.NewBufferString("")
	}

	// create the http request, use the builtin context's context to ensure
	// the request is cancelled if evaluation is cancelled.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(bctx.Context)

	// Add custom headers
	if len(customHeaders) != 0 {
		if ok, err := addHeaders(req, customHeaders); !ok {
			return nil, err
		}
		// Don't overwrite or append to one that was set in the custom headers
		if _, hasUA := customHeaders["User-Agent"]; !hasUA {
			req.Header.Add("User-Agent", version.UserAgent)
		}
	}

	// execute the http request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// format the http result
	var resultBody interface{}
	var resultRawBody []byte

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	resultRawBody, err = ioutil.ReadAll(tee)
	if err != nil {
		return nil, err
	}

	// If the response body cannot be JSON decoded,
	// an error will not be returned. Instead the "body" field
	// in the result will be null.
	if isContentTypeJSON(resp.Header) || forceJSONDecode {
		json.NewDecoder(&buf).Decode(&resultBody)
	}

	result := make(map[string]interface{})
	result["status"] = resp.Status
	result["status_code"] = resp.StatusCode
	result["body"] = resultBody
	result["raw_body"] = string(resultRawBody)

	resultObj, err := ast.InterfaceToValue(result)
	if err != nil {
		return nil, err
	}

	return resultObj, nil
}

func isContentTypeJSON(header http.Header) bool {
	return strings.Contains(header.Get("Content-Type"), "application/json")
}

// In the BuiltinContext cache we only store a single entry that points to
// our ValueMap which is the "real" http.send() cache.
func getHTTPSendCache(bctx BuiltinContext) *ast.ValueMap {
	raw, ok := bctx.Cache.Get(httpSendBuiltinCacheKey)
	if !ok {
		// Initialize if it isn't there
		cache := ast.NewValueMap()
		bctx.Cache.Put(httpSendBuiltinCacheKey, cache)
		return cache
	}

	cache, ok := raw.(*ast.ValueMap)
	if !ok {
		return nil
	}
	return cache
}

// checkHTTPSendCache checks for the given key's value in the cache
func checkHTTPSendCache(bctx BuiltinContext, key ast.Object) ast.Value {
	requestCache := getHTTPSendCache(bctx)
	if requestCache == nil {
		return nil
	}

	return requestCache.Get(key)
}

func insertIntoHTTPSendCache(bctx BuiltinContext, key ast.Object, value ast.Value) {
	requestCache := getHTTPSendCache(bctx)
	if requestCache == nil {
		// Should never happen.. if it does just skip caching the value
		return
	}
	requestCache.Put(key, value)
}

func createAllowedKeys() {
	for _, element := range allowedKeyNames {
		allowedKeys.Add(ast.StringTerm(element))
	}
}

func parseTimeout(timeoutVal ast.Value) (time.Duration, error) {
	var timeout time.Duration
	switch t := timeoutVal.(type) {
	case ast.Number:
		timeoutInt, ok := t.Int64()
		if !ok {
			return timeout, fmt.Errorf("invalid timeout number value %v, must be int64", timeoutVal)
		}
		return time.Duration(timeoutInt), nil
	case ast.String:
		// Support strings without a unit, treat them the same as just a number value (ns)
		var err error
		timeoutInt, err := strconv.ParseInt(string(t), 10, 64)
		if err == nil {
			return time.Duration(timeoutInt), nil
		}

		// Try parsing it as a duration (requires a supported units suffix)
		timeout, err = time.ParseDuration(string(t))
		if err != nil {
			return timeout, fmt.Errorf("invalid timeout value %v: %s", timeoutVal, err)
		}
		return timeout, nil
	default:
		return timeout, builtins.NewOperandErr(1, "'timeout' must be one of {string, number} but got %s", ast.TypeName(t))
	}
}
