// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/version"
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
	"tls_ca_cert",
	"tls_ca_cert_file",
	"tls_ca_cert_env_variable",
	"tls_client_cert",
	"tls_client_cert_file",
	"tls_client_cert_env_variable",
	"tls_client_key",
	"tls_client_key_file",
	"tls_client_key_env_variable",
	"tls_insecure_skip_verify",
	"tls_server_name",
	"timeout",
	"cache",
	"force_cache",
	"force_cache_duration_seconds",
	"raise_error",
}

var (
	allowedKeys              = ast.NewSet()
	requiredKeys             = ast.NewSet(ast.StringTerm("method"), ast.StringTerm("url"))
	httpSendLatencyMetricKey = "rego_builtin_" + strings.ReplaceAll(ast.HTTPSend.Name, ".", "_")
)

type httpSendKey string

const (
	// httpSendBuiltinCacheKey is the key in the builtin context cache that
	// points to the http.send() specific cache resides at.
	httpSendBuiltinCacheKey httpSendKey = "HTTP_SEND_CACHE_KEY"

	// HTTPSendInternalErr represents a runtime evaluation error.
	HTTPSendInternalErr string = "eval_http_send_internal_error"

	// HTTPSendNetworkErr represents a network error.
	HTTPSendNetworkErr string = "eval_http_send_network_error"
)

func builtinHTTPSend(bctx BuiltinContext, args []*ast.Term, iter func(*ast.Term) error) error {
	req, err := validateHTTPRequestOperand(args[0], 1)
	if err != nil {
		return handleBuiltinErr(ast.HTTPSend.Name, bctx.Location, err)
	}

	raiseError, err := getRaiseErrorValue(req)
	if err != nil {
		return handleBuiltinErr(ast.HTTPSend.Name, bctx.Location, err)
	}

	result, err := getHTTPResponse(bctx, req)
	if err != nil {
		if raiseError {
			return handleHTTPSendErr(bctx, err)
		}

		obj := ast.NewObject()
		obj.Insert(ast.StringTerm("status_code"), ast.IntNumberTerm(0))

		errObj := ast.NewObject()

		switch err.(type) {
		case *url.Error:
			errObj.Insert(ast.StringTerm("code"), ast.StringTerm(HTTPSendNetworkErr))
		default:
			errObj.Insert(ast.StringTerm("code"), ast.StringTerm(HTTPSendInternalErr))
		}

		errObj.Insert(ast.StringTerm("message"), ast.StringTerm(err.Error()))
		obj.Insert(ast.StringTerm("error"), ast.NewTerm(errObj))

		result = ast.NewTerm(obj)
	}
	return iter(result)
}

func getHTTPResponse(bctx BuiltinContext, req ast.Object) (*ast.Term, error) {

	bctx.Metrics.Timer(httpSendLatencyMetricKey).Start()

	reqExecutor, err := newHTTPRequestExecutor(bctx, req)
	if err != nil {
		return nil, err
	}

	// check if cache already has a response for this query
	resp, err := reqExecutor.CheckCache()
	if err != nil {
		return nil, err
	}

	if resp == nil {
		httpResp, err := reqExecutor.ExecuteHTTPRequest()
		if err != nil {
			return nil, err
		}

		// add result to cache
		resp, err = reqExecutor.InsertIntoCache(httpResp)
		if err != nil {
			return nil, err
		}
	}

	bctx.Metrics.Timer(httpSendLatencyMetricKey).Stop()

	return ast.NewTerm(resp), nil
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

// canonicalizeHeaders returns a copy of the headers where the keys are in
// canonical HTTP form.
func canonicalizeHeaders(headers map[string]interface{}) map[string]interface{} {
	canonicalized := map[string]interface{}{}

	for k, v := range headers {
		canonicalized[http.CanonicalHeaderKey(k)] = v
	}

	return canonicalized
}

func createHTTPRequest(bctx BuiltinContext, obj ast.Object) (*http.Request, *http.Client, error) {
	var url string
	var method string

	// Additional CA certificates loading options.
	var tlsCaCert []byte
	var tlsCaCertEnvVar string
	var tlsCaCertFile string

	// Client TLS certificate and key options. Each input source
	// comes in a matched pair.
	var tlsClientCert []byte
	var tlsClientKey []byte

	var tlsClientCertEnvVar string
	var tlsClientKeyEnvVar string

	var tlsClientCertFile string
	var tlsClientKeyFile string

	var tlsServerName string
	var body *bytes.Buffer
	var rawBody *bytes.Buffer
	var enableRedirect bool
	var tlsUseSystemCerts bool
	var tlsConfig tls.Config
	var customHeaders map[string]interface{}
	var tlsInsecureSkipVerify bool
	var timeout = defaultHTTPRequestTimeout

	for _, val := range obj.Keys() {
		key, err := ast.JSON(val.Value)
		if err != nil {
			return nil, nil, err
		}

		key = key.(string)

		var strVal string

		if s, ok := obj.Get(val).Value.(ast.String); ok {
			strVal = string(s)
		} else {
			// Most parameters are strings, so consolidate the type checking.
			switch key {
			case "method",
				"url",
				"raw_body",
				"tls_ca_cert",
				"tls_ca_cert_file",
				"tls_ca_cert_env_variable",
				"tls_client_cert",
				"tls_client_cert_file",
				"tls_client_cert_env_variable",
				"tls_client_key",
				"tls_client_key_file",
				"tls_client_key_env_variable",
				"tls_server_name":
				return nil, nil, fmt.Errorf("%q must be a string", key)
			}
		}

		switch key {
		case "method":
			method = strings.ToUpper(strings.Trim(strVal, "\""))
		case "url":
			url = strings.Trim(strVal, "\"")
		case "enable_redirect":
			enableRedirect, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, nil, err
			}
		case "body":
			bodyVal := obj.Get(val).Value
			bodyValInterface, err := ast.JSON(bodyVal)
			if err != nil {
				return nil, nil, err
			}

			bodyValBytes, err := json.Marshal(bodyValInterface)
			if err != nil {
				return nil, nil, err
			}
			body = bytes.NewBuffer(bodyValBytes)
		case "raw_body":
			rawBody = bytes.NewBuffer([]byte(strVal))
		case "tls_use_system_certs":
			tlsUseSystemCerts, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, nil, err
			}
		case "tls_ca_cert":
			tlsCaCert = bytes.Trim([]byte(strVal), "\"")
		case "tls_ca_cert_file":
			tlsCaCertFile = strings.Trim(strVal, "\"")
		case "tls_ca_cert_env_variable":
			tlsCaCertEnvVar = strings.Trim(strVal, "\"")
		case "tls_client_cert":
			tlsClientCert = bytes.Trim([]byte(strVal), "\"")
		case "tls_client_cert_file":
			tlsClientCertFile = strings.Trim(strVal, "\"")
		case "tls_client_cert_env_variable":
			tlsClientCertEnvVar = strings.Trim(strVal, "\"")
		case "tls_client_key":
			tlsClientKey = bytes.Trim([]byte(strVal), "\"")
		case "tls_client_key_file":
			tlsClientKeyFile = strings.Trim(strVal, "\"")
		case "tls_client_key_env_variable":
			tlsClientKeyEnvVar = strings.Trim(strVal, "\"")
		case "tls_server_name":
			tlsServerName = strings.Trim(strVal, "\"")
		case "headers":
			headersVal := obj.Get(val).Value
			headersValInterface, err := ast.JSON(headersVal)
			if err != nil {
				return nil, nil, err
			}
			var ok bool
			customHeaders, ok = headersValInterface.(map[string]interface{})
			if !ok {
				return nil, nil, fmt.Errorf("invalid type for headers key")
			}
		case "tls_insecure_skip_verify":
			tlsInsecureSkipVerify, err = strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, nil, err
			}
		case "timeout":
			timeout, err = parseTimeout(obj.Get(val).Value)
			if err != nil {
				return nil, nil, err
			}
		case "cache", "force_cache", "force_cache_duration_seconds", "force_json_decode", "raise_error": // no-op
		default:
			return nil, nil, fmt.Errorf("invalid parameter %q", key)
		}
	}

	isTLS := false
	client := &http.Client{
		Timeout: timeout,
	}

	if tlsInsecureSkipVerify {
		isTLS = true
		tlsConfig.InsecureSkipVerify = tlsInsecureSkipVerify
	}

	if len(tlsClientCert) > 0 && len(tlsClientKey) > 0 {
		cert, err := tls.X509KeyPair(tlsClientCert, tlsClientKey)
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	if tlsClientCertFile != "" && tlsClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsClientCertFile, tlsClientKeyFile)
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	if tlsClientCertEnvVar != "" && tlsClientKeyEnvVar != "" {
		cert, err := tls.X509KeyPair(
			[]byte(os.Getenv(tlsClientCertEnvVar)),
			[]byte(os.Getenv(tlsClientKeyEnvVar)))
		if err != nil {
			return nil, nil, fmt.Errorf("cannot extract public/private key pair from envvars %q, %q: %w",
				tlsClientCertEnvVar, tlsClientKeyEnvVar, err)
		}

		isTLS = true
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	// Check the system certificates config first so that we
	// load additional certificated into the correct pool.
	if tlsUseSystemCerts {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.RootCAs = pool
	}

	if len(tlsCaCert) != 0 {
		tlsCaCert = bytes.Replace(tlsCaCert, []byte("\\n"), []byte("\n"), -1)
		pool, err := addCACertsFromBytes(tlsConfig.RootCAs, []byte(tlsCaCert))
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.RootCAs = pool
	}

	if tlsCaCertFile != "" {
		pool, err := addCACertsFromFile(tlsConfig.RootCAs, tlsCaCertFile)
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.RootCAs = pool
	}

	if tlsCaCertEnvVar != "" {
		pool, err := addCACertsFromEnv(tlsConfig.RootCAs, tlsCaCertEnvVar)
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.RootCAs = pool
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
		return nil, nil, err
	}

	req = req.WithContext(bctx.Context)

	// Add custom headers
	if len(customHeaders) != 0 {
		customHeaders = canonicalizeHeaders(customHeaders)

		for k, v := range customHeaders {
			header, ok := v.(string)
			if !ok {
				return nil, nil, fmt.Errorf("invalid type for headers value %q", v)
			}

			req.Header.Add(k, header)
		}

		// Don't overwrite or append to one that was set in the custom headers
		if _, hasUA := customHeaders["User-Agent"]; !hasUA {
			req.Header.Add("User-Agent", version.UserAgent)
		}

		// If the caller specifies the Host header, use it for the HTTP
		// request host and the TLS server name.
		if host, hasHost := customHeaders["Host"]; hasHost {
			host := host.(string) // We already checked that it's a string.
			req.Host = host

			// Only default the ServerName if the caller has
			// specified the host. If we don't specify anything,
			// Go will default to the target hostname. This name
			// is not the same as the default that Go populates
			// `req.Host` with, which is why we don't just set
			// this unconditionally.
			tlsConfig.ServerName = host
		}
	}

	if tlsServerName != "" {
		tlsConfig.ServerName = tlsServerName
	}

	return req, client, nil
}

func executeHTTPRequest(req *http.Request, client *http.Client) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
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

// checkHTTPSendInterQueryCache checks for the given key's value in the inter-query cache
func checkHTTPSendInterQueryCache(bctx BuiltinContext, key ast.Object, req *http.Request, client *http.Client, forceJSONDecode bool, cacheParams *forceCacheParams) (ast.Value, error) {
	requestCache := bctx.InterQueryBuiltinCache

	value, found := requestCache.Get(key)
	if !found {
		return nil, nil
	}

	cachedResp, ok := value.(*interQueryCacheValue)
	if !ok {
		return nil, nil
	}

	// check the freshness of the cached response
	if isCachedResponseFresh(bctx, cachedResp, cacheParams) {
		return cachedResp.value, nil
	}

	// check with the server if the stale response is still up-to-date.
	// If server returns a new response (ie. status_code=200), update the cache with the new response
	// If server returns an unmodified response (ie. status_code=304), update the headers for the existing response
	result, modified, err := revalidateCachedResponse(req, client, cachedResp)
	requestCache.Delete(key)
	if err != nil || result == nil {
		return nil, err
	}

	defer result.Body.Close()

	var newValue ast.Value
	var size int

	if !modified {
		// update the headers in the cached response with their corresponding values from the 304 (Not Modified) response
		cachedRespObj := cachedResp.value.(ast.Object)
		existingHeaders := cachedRespObj.Get(ast.StringTerm("headers"))
		existingHeadersObj := existingHeaders.Value.(ast.Object)

		newHeaders := getResponseHeaders(result.Header)
		for k, v := range newHeaders {
			valueAST, err := ast.InterfaceToValue(v)
			if err != nil {
				return nil, err
			}
			existingHeadersObj.Insert(ast.StringTerm(k), ast.NewTerm(valueAST))
		}

		cachedRespObj.Insert(ast.StringTerm("headers"), ast.NewTerm(existingHeadersObj))
		newValue = cachedRespObj
		size = cachedResp.size
	} else {
		newValue, size, err = formatHTTPResponseToAST(result, forceJSONDecode)
		if err != nil {
			return nil, err
		}
	}

	err = insertIntoHTTPSendInterQueryCache(bctx, key, newValue, result, size)
	if err != nil {
		return nil, err
	}
	return newValue, nil
}

// insertIntoHTTPSendInterQueryCache inserts given key and value in the inter-query cache
func insertIntoHTTPSendInterQueryCache(bctx BuiltinContext, key, value ast.Value, resp *http.Response, size int) error {
	if resp == nil || !canStore(resp.Header) {
		return nil
	}

	requestCache := bctx.InterQueryBuiltinCache

	pcv, err := newInterQueryCacheValue(resp, value, size)
	if err != nil {
		return err
	}

	requestCache.Insert(key, pcv)
	return nil
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

func getBoolValFromReqObj(req ast.Object, key *ast.Term) (bool, error) {
	var b ast.Boolean
	var ok bool
	if v := req.Get(key); v != nil {
		if b, ok = v.Value.(ast.Boolean); !ok {
			return false, fmt.Errorf("invalid value for %v field", key.String())
		}
	}
	return bool(b), nil
}

type interQueryCacheValue struct {
	value        ast.Value         // http response
	size         int               // size of response body
	date         time.Time         // origination date and time of response
	cacheControl map[string]string // response cache-control header
	maxAge       deltaSeconds      // max-age cache control directive
	expires      time.Time         // date/time after which the response is considered stale
	etag         string            // identifier for a specific version of the response
	lastModified string            // date and time response was last modified as per origin server
}

// deltaSeconds specifies a non-negative integer, representing
// time in seconds: http://tools.ietf.org/html/rfc7234#section-1.2.1
type deltaSeconds int32

func newInterQueryCacheValue(resp *http.Response, value ast.Value, size int) (*interQueryCacheValue, error) {
	cv := interQueryCacheValue{value: value, size: size, maxAge: -1}

	err := parseResponseAndInjectHeaders(resp, &cv)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func (c interQueryCacheValue) SizeInBytes() int64 {
	return int64(c.size)
}

func parseResponseAndInjectHeaders(resp *http.Response, pcv *interQueryCacheValue) error {
	var err error

	pcv.date, err = getResponseHeaderDate(resp.Header)
	if err != nil {
		return err
	}

	pcv.cacheControl = parseCacheControlHeader(resp.Header)
	pcv.maxAge, err = parseMaxAgeCacheDirective(pcv.cacheControl)
	if err != nil {
		return err
	}

	pcv.expires, err = getResponseHeaderExpires(resp.Header)
	if err != nil {
		return err
	}

	pcv.etag = resp.Header.Get("etag")

	pcv.lastModified = resp.Header.Get("last-modified")

	return nil
}

func revalidateCachedResponse(req *http.Request, client *http.Client, resp *interQueryCacheValue) (*http.Response, bool, error) {
	etag := resp.etag
	lastModified := resp.lastModified

	if etag == "" && lastModified == "" {
		return nil, false, nil
	}

	cloneReq := req.Clone(req.Context())

	if etag != "" {
		cloneReq.Header.Set("if-none-match", etag)
	}

	if lastModified != "" {
		cloneReq.Header.Set("if-modified-since", lastModified)
	}

	response, err := client.Do(cloneReq)
	if err != nil {
		return nil, false, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return response, true, nil

	case http.StatusNotModified:
		return response, false, nil
	}

	return nil, false, nil
}

func canStore(headers http.Header) bool {
	ccHeaders := parseCacheControlHeader(headers)

	// Check "no-store" cache directive
	// The "no-store" response directive indicates that a cache MUST NOT
	// store any part of either the immediate request or response.
	if _, ok := ccHeaders["no-store"]; ok {
		return false
	}
	return true
}

func isCachedResponseFresh(bctx BuiltinContext, resp *interQueryCacheValue, cacheParams *forceCacheParams) bool {
	if resp.date.IsZero() {
		return false
	}

	currentTime := getCurrentTime(bctx)
	if currentTime.IsZero() {
		return false
	}

	currentAge := currentTime.Sub(resp.date)

	// The time.Sub operation uses wall clock readings and
	// not monotonic clock readings as the parsed version of the response time
	// does not contain monotonic clock readings. This can result in negative durations.
	// Another scenario where a negative duration can occur, is when a server sets the Date
	// response header. As per https://tools.ietf.org/html/rfc7231#section-7.1.1.2,
	// an origin server MUST NOT send a Date header field if it does not
	// have a clock capable of providing a reasonable approximation of the
	// current instance in Coordinated Universal Time.
	// Hence, consider the cached response as stale if a negative duration is encountered.
	if currentAge < 0 {
		return false
	}

	if cacheParams != nil {
		// override the cache directives set by the server
		maxAgeDur := time.Second * time.Duration(cacheParams.forceCacheDurationSeconds)
		if maxAgeDur > currentAge {
			return true
		}
	} else {
		// Check "max-age" cache directive.
		// The "max-age" response directive indicates that the response is to be
		// considered stale after its age is greater than the specified number
		// of seconds.
		if resp.maxAge != -1 {
			maxAgeDur := time.Second * time.Duration(resp.maxAge)
			if maxAgeDur > currentAge {
				return true
			}
		} else {
			// Check "Expires" header.
			// Note: "max-age" if set, takes precedence over "Expires"
			if resp.expires.Sub(resp.date) > currentAge {
				return true
			}
		}
	}
	return false
}

func getCurrentTime(bctx BuiltinContext) time.Time {
	var current time.Time

	value, err := ast.JSON(bctx.Time.Value)
	if err != nil {
		return current
	}

	valueNum, ok := value.(json.Number)
	if !ok {
		return current
	}

	valueNumInt, err := valueNum.Int64()
	if err != nil {
		return current
	}

	current = time.Unix(0, valueNumInt).UTC()
	return current
}

func parseCacheControlHeader(headers http.Header) map[string]string {
	ccDirectives := map[string]string{}
	ccHeader := headers.Get("cache-control")

	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			items := strings.Split(part, "=")
			if len(items) != 2 {
				continue
			}
			ccDirectives[strings.Trim(items[0], " ")] = strings.Trim(items[1], ",")
		} else {
			ccDirectives[part] = ""
		}
	}

	return ccDirectives
}

func getResponseHeaderDate(headers http.Header) (date time.Time, err error) {
	dateHeader := headers.Get("date")
	if dateHeader == "" {
		err = fmt.Errorf("no date header")
		return
	}
	return time.Parse(time.RFC1123, dateHeader)
}

func getResponseHeaderExpires(headers http.Header) (date time.Time, err error) {
	expiresHeader := headers.Get("expires")
	if expiresHeader == "" {
		return
	}
	return time.Parse(time.RFC1123, expiresHeader)
}

// parseMaxAgeCacheDirective parses the max-age directive expressed in delta-seconds as per
// https://tools.ietf.org/html/rfc7234#section-1.2.1
func parseMaxAgeCacheDirective(cc map[string]string) (deltaSeconds, error) {
	maxAge, ok := cc["max-age"]
	if !ok {
		return deltaSeconds(-1), nil
	}

	val, err := strconv.ParseUint(maxAge, 10, 32)
	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				return deltaSeconds(math.MaxInt32), nil
			}
		}
		return deltaSeconds(-1), err
	}

	if val > math.MaxInt32 {
		return deltaSeconds(math.MaxInt32), nil
	}
	return deltaSeconds(val), nil
}

func formatHTTPResponseToAST(resp *http.Response, forceJSONDecode bool) (ast.Value, int, error) {

	var resultBody interface{}
	var resultRawBody []byte

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	resultRawBody, err := ioutil.ReadAll(tee)
	if err != nil {
		return nil, 0, err
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
	result["headers"] = getResponseHeaders(resp.Header)

	resultObj, err := ast.InterfaceToValue(result)
	if err != nil {
		return nil, 0, err
	}

	return resultObj, len(resultRawBody), nil
}

func getResponseHeaders(headers http.Header) map[string]interface{} {
	respHeaders := map[string]interface{}{}
	for headerName, values := range headers {
		var respValues []interface{}
		for _, v := range values {
			respValues = append(respValues, v)
		}
		respHeaders[strings.ToLower(headerName)] = respValues
	}
	return respHeaders
}

// httpRequestExecutor defines an interface for the http send cache
type httpRequestExecutor interface {
	CheckCache() (ast.Value, error)
	InsertIntoCache(value *http.Response) (ast.Value, error)
	ExecuteHTTPRequest() (*http.Response, error)
}

// newHTTPRequestExecutor returns a new HTTP request executor that wraps either an inter-query or
// intra-query cache implementation
func newHTTPRequestExecutor(bctx BuiltinContext, key ast.Object) (httpRequestExecutor, error) {
	useInterQueryCache, forceCacheParams, err := useInterQueryCache(key)
	if err != nil {
		return nil, handleHTTPSendErr(bctx, err)
	}

	if useInterQueryCache && bctx.InterQueryBuiltinCache != nil {
		return newInterQueryCache(bctx, key, forceCacheParams)
	}
	return newIntraQueryCache(bctx, key)
}

type interQueryCache struct {
	bctx             BuiltinContext
	key              ast.Object
	httpReq          *http.Request
	httpClient       *http.Client
	forceJSONDecode  bool
	forceCacheParams *forceCacheParams
}

func newInterQueryCache(bctx BuiltinContext, key ast.Object, forceCacheParams *forceCacheParams) (*interQueryCache, error) {
	return &interQueryCache{bctx: bctx, key: key, forceCacheParams: forceCacheParams}, nil
}

// CheckCache checks the cache for the value of the key set on this object
func (c *interQueryCache) CheckCache() (ast.Value, error) {
	var err error

	c.httpReq, c.httpClient, err = createHTTPRequest(c.bctx, c.key)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	c.forceJSONDecode, err = getBoolValFromReqObj(c.key, ast.StringTerm("force_json_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	resp, err := checkHTTPSendInterQueryCache(c.bctx, c.key, c.httpReq, c.httpClient, c.forceJSONDecode, c.forceCacheParams)

	// fallback to the intra-query cache if response not found in the inter-query cache or inter-query cache look-up results
	// in an error
	if resp == nil || err != nil {
		return checkHTTPSendCache(c.bctx, c.key), nil
	}
	return resp, err
}

// InsertIntoCache inserts the key set on this object into the cache with the given value
func (c *interQueryCache) InsertIntoCache(value *http.Response) (ast.Value, error) {
	result, size, err := formatHTTPResponseToAST(value, c.forceJSONDecode)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	// fallback to the intra-query cache if error encountered while inserting response in inter-query cache
	err = insertIntoHTTPSendInterQueryCache(c.bctx, c.key, result, value, size)
	if err != nil {
		insertIntoHTTPSendCache(c.bctx, c.key, result)
	}
	return result, nil
}

// ExecuteHTTPRequest executes a HTTP request
func (c *interQueryCache) ExecuteHTTPRequest() (*http.Response, error) {
	return executeHTTPRequest(c.httpReq, c.httpClient)
}

type intraQueryCache struct {
	bctx BuiltinContext
	key  ast.Object
}

func newIntraQueryCache(bctx BuiltinContext, key ast.Object) (*intraQueryCache, error) {
	return &intraQueryCache{bctx: bctx, key: key}, nil
}

// CheckCache checks the cache for the value of the key set on this object
func (c *intraQueryCache) CheckCache() (ast.Value, error) {
	return checkHTTPSendCache(c.bctx, c.key), nil
}

// InsertIntoCache inserts the key set on this object into the cache with the given value
func (c *intraQueryCache) InsertIntoCache(value *http.Response) (ast.Value, error) {
	forceJSONDecode, err := getBoolValFromReqObj(c.key, ast.StringTerm("force_json_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	result, _, err := formatHTTPResponseToAST(value, forceJSONDecode)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	insertIntoHTTPSendCache(c.bctx, c.key, result)
	return result, nil
}

// ExecuteHTTPRequest executes a HTTP request
func (c *intraQueryCache) ExecuteHTTPRequest() (*http.Response, error) {
	httpReq, httpClient, err := createHTTPRequest(c.bctx, c.key)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}
	return executeHTTPRequest(httpReq, httpClient)
}

func useInterQueryCache(req ast.Object) (bool, *forceCacheParams, error) {
	value, err := getBoolValFromReqObj(req, ast.StringTerm("cache"))
	if err != nil {
		return false, nil, err
	}

	valueForceCache, err := getBoolValFromReqObj(req, ast.StringTerm("force_cache"))
	if err != nil {
		return false, nil, err
	}

	if valueForceCache {
		forceCacheParams, err := newForceCacheParams(req)
		return true, forceCacheParams, err
	}

	return value, nil, nil
}

type forceCacheParams struct {
	forceCacheDurationSeconds int32
}

func newForceCacheParams(req ast.Object) (*forceCacheParams, error) {
	term := req.Get(ast.StringTerm("force_cache_duration_seconds"))
	if term == nil {
		return nil, fmt.Errorf("'force_cache' set but 'force_cache_duration_seconds' parameter is missing")
	}

	forceCacheDurationSeconds := term.String()

	value, err := strconv.ParseInt(forceCacheDurationSeconds, 10, 32)
	if err != nil {
		return nil, err
	}

	return &forceCacheParams{forceCacheDurationSeconds: int32(value)}, nil
}

func getRaiseErrorValue(req ast.Object) (bool, error) {
	result := ast.Boolean(true)
	var ok bool
	if v := req.Get(ast.StringTerm("raise_error")); v != nil {
		if result, ok = v.Value.(ast.Boolean); !ok {
			return false, fmt.Errorf("invalid value for raise_error field")
		}
	}
	return bool(result), nil
}
