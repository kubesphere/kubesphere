// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/version"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/topdown/cache"
	"github.com/open-policy-agent/opa/tracing"
	"github.com/open-policy-agent/opa/util"
)

type cachingMode string

const (
	defaultHTTPRequestTimeoutEnv             = "HTTP_SEND_TIMEOUT"
	defaultCachingMode           cachingMode = "serialized"
	cachingModeDeserialized      cachingMode = "deserialized"
)

var defaultHTTPRequestTimeout = time.Second * 5

var allowedKeyNames = [...]string{
	"method",
	"url",
	"body",
	"enable_redirect",
	"force_json_decode",
	"force_yaml_decode",
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
	"caching_mode",
	"max_retry_attempts",
}

// ref: https://www.rfc-editor.org/rfc/rfc7231#section-6.1
var cacheableHTTPStatusCodes = [...]int{
	http.StatusOK,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
	http.StatusPartialContent,
	http.StatusMultipleChoices,
	http.StatusMovedPermanently,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusGone,
	http.StatusRequestURITooLong,
	http.StatusNotImplemented,
}

var (
	allowedKeys                 = ast.NewSet()
	cacheableCodes              = ast.NewSet()
	requiredKeys                = ast.NewSet(ast.StringTerm("method"), ast.StringTerm("url"))
	httpSendLatencyMetricKey    = "rego_builtin_" + strings.ReplaceAll(ast.HTTPSend.Name, ".", "_")
	httpSendInterQueryCacheHits = httpSendLatencyMetricKey + "_interquery_cache_hits"
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

	// minRetryDelay is amount of time to backoff after the first failure.
	minRetryDelay = time.Millisecond * 100

	// maxRetryDelay is the upper bound of backoff delay.
	maxRetryDelay = time.Second * 60
)

func builtinHTTPSend(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	req, err := validateHTTPRequestOperand(operands[0], 1)
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

	// Check if cache already has a response for this query
	resp, err := reqExecutor.CheckCache()
	if err != nil {
		return nil, err
	}

	if resp == nil {
		httpResp, err := reqExecutor.ExecuteHTTPRequest()
		if err != nil {
			reqExecutor.InsertErrorIntoCache(err)
			return nil, err
		}
		defer util.Close(httpResp)
		// Add result to intra/inter-query cache.
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
	createCacheableHTTPStatusCodes()
	initDefaults()
	RegisterBuiltinFunc(ast.HTTPSend.Name, builtinHTTPSend)
}

func handleHTTPSendErr(bctx BuiltinContext, err error) error {
	// Return HTTP client timeout errors in a generic error message to avoid confusion about what happened.
	// Do not do this if the builtin context was cancelled and is what caused the request to stop.
	if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() && bctx.Context.Err() == nil {
		err = fmt.Errorf("%s %s: request timed out", urlErr.Op, urlErr.URL)
	}
	if err := bctx.Context.Err(); err != nil {
		return Halt{
			Err: &Error{
				Code:    CancelErr,
				Message: fmt.Sprintf("http.send: timed out (%s)", err.Error()),
			},
		}
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

// useSocket examines the url for "unix://" and returns a *http.Transport with
// a DialContext that opens a socket (specified in the http call).
// The url is expected to contain socket=/path/to/socket (url encoded)
// Ex. "unix://localhost/end/point?socket=%2Ftmp%2Fhttp.sock"
func useSocket(rawURL string, tlsConfig *tls.Config) (bool, string, *http.Transport) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, "", nil
	}

	if u.Scheme != "unix" || u.RawQuery == "" {
		return false, rawURL, nil
	}

	v, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return false, rawURL, nil
	}

	// Rewrite URL targeting the UNIX domain socket.
	u.Scheme = "http"

	// Extract the path to the socket.
	// Only retrieve the first value. Subsequent values are ignored and removed
	// to prevent HTTP parameter pollution.
	socket := v.Get("socket")
	v.Del("socket")
	u.RawQuery = v.Encode()

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return http.DefaultTransport.(*http.Transport).DialContext(ctx, "unix", socket)
	}
	tr.TLSClientConfig = tlsConfig
	tr.DisableKeepAlives = true

	return true, u.String(), tr
}

func verifyHost(bctx BuiltinContext, host string) error {
	if bctx.Capabilities == nil || bctx.Capabilities.AllowNet == nil {
		return nil
	}

	for _, allowed := range bctx.Capabilities.AllowNet {
		if allowed == host {
			return nil
		}
	}

	return fmt.Errorf("unallowed host: %s", host)
}

func verifyURLHost(bctx BuiltinContext, unverifiedURL string) error {
	// Eager return to avoid unnecessary URL parsing
	if bctx.Capabilities == nil || bctx.Capabilities.AllowNet == nil {
		return nil
	}

	parsedURL, err := url.Parse(unverifiedURL)
	if err != nil {
		return err
	}

	host := strings.Split(parsedURL.Host, ":")[0]

	return verifyHost(bctx, host)
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
	var tlsUseSystemCerts *bool
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
			strVal = strings.Trim(string(s), "\"")
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
			method = strings.ToUpper(strVal)
		case "url":
			err := verifyURLHost(bctx, strVal)
			if err != nil {
				return nil, nil, err
			}
			url = strVal
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
			tempTLSUseSystemCerts, err := strconv.ParseBool(obj.Get(val).String())
			if err != nil {
				return nil, nil, err
			}
			tlsUseSystemCerts = &tempTLSUseSystemCerts
		case "tls_ca_cert":
			tlsCaCert = []byte(strVal)
		case "tls_ca_cert_file":
			tlsCaCertFile = strVal
		case "tls_ca_cert_env_variable":
			tlsCaCertEnvVar = strVal
		case "tls_client_cert":
			tlsClientCert = []byte(strVal)
		case "tls_client_cert_file":
			tlsClientCertFile = strVal
		case "tls_client_cert_env_variable":
			tlsClientCertEnvVar = strVal
		case "tls_client_key":
			tlsClientKey = []byte(strVal)
		case "tls_client_key_file":
			tlsClientKeyFile = strVal
		case "tls_client_key_env_variable":
			tlsClientKeyEnvVar = strVal
		case "tls_server_name":
			tlsServerName = strVal
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
		case "cache", "caching_mode",
			"force_cache", "force_cache_duration_seconds",
			"force_json_decode", "force_yaml_decode",
			"raise_error", "max_retry_attempts": // no-op
		default:
			return nil, nil, fmt.Errorf("invalid parameter %q", key)
		}
	}

	isTLS := false
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
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

	// Use system certs if no CA cert is provided
	// or system certs flag is not set
	if len(tlsCaCert) == 0 && tlsCaCertFile == "" && tlsCaCertEnvVar == "" && tlsUseSystemCerts == nil {
		trueValue := true
		tlsUseSystemCerts = &trueValue
	}

	// Check the system certificates config first so that we
	// load additional certificated into the correct pool.
	if tlsUseSystemCerts != nil && *tlsUseSystemCerts && runtime.GOOS != "windows" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, nil, err
		}

		isTLS = true
		tlsConfig.RootCAs = pool
	}

	if len(tlsCaCert) != 0 {
		tlsCaCert = bytes.Replace(tlsCaCert, []byte("\\n"), []byte("\n"), -1)
		pool, err := addCACertsFromBytes(tlsConfig.RootCAs, tlsCaCert)
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
		if ok, parsedURL, tr := useSocket(url, &tlsConfig); ok {
			client.Transport = tr
			url = parsedURL
		} else {
			tr := http.DefaultTransport.(*http.Transport).Clone()
			tr.TLSClientConfig = &tlsConfig
			tr.DisableKeepAlives = true
			client.Transport = tr
		}
	} else {
		if ok, parsedURL, tr := useSocket(url, nil); ok {
			client.Transport = tr
			url = parsedURL
		}
	}

	// check if redirects are enabled
	if enableRedirect {
		client.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
			return verifyURLHost(bctx, req.URL.String())
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

	if len(bctx.DistributedTracingOpts) > 0 {
		client.Transport = tracing.NewTransport(client.Transport, bctx.DistributedTracingOpts)
	}

	return req, client, nil
}

func executeHTTPRequest(req *http.Request, client *http.Client, inputReqObj ast.Object) (*http.Response, error) {
	var err error
	var retry int

	retry, err = getNumberValFromReqObj(inputReqObj, ast.StringTerm("max_retry_attempts"))
	if err != nil {
		return nil, err
	}

	for i := 0; true; i++ {

		var resp *http.Response
		resp, err = client.Do(req)
		if err == nil {
			return resp, nil
		}

		// final attempt
		if i == retry {
			break
		}

		if err == context.Canceled {
			return nil, err
		}

		select {
		case <-time.After(util.DefaultBackoff(float64(minRetryDelay), float64(maxRetryDelay), i)):
		case <-req.Context().Done():
			return nil, context.Canceled
		}
	}
	return nil, err
}

func isContentType(header http.Header, typ ...string) bool {
	for _, t := range typ {
		if strings.Contains(header.Get("Content-Type"), t) {
			return true
		}
	}
	return false
}

type httpSendCacheEntry struct {
	response *ast.Value
	error    error
}

// The httpSendCache is used for intra-query caching of http.send results.
type httpSendCache struct {
	entries *util.HashMap
}

func newHTTPSendCache() *httpSendCache {
	return &httpSendCache{
		entries: util.NewHashMap(valueEq, valueHash),
	}
}

func valueHash(v util.T) int {
	return v.(ast.Value).Hash()
}

func valueEq(a, b util.T) bool {
	av := a.(ast.Value)
	bv := b.(ast.Value)
	return av.Compare(bv) == 0
}

func (cache *httpSendCache) get(k ast.Value) *httpSendCacheEntry {
	if v, ok := cache.entries.Get(k); ok {
		v := v.(httpSendCacheEntry)
		return &v
	}
	return nil
}

func (cache *httpSendCache) putResponse(k ast.Value, v *ast.Value) {
	cache.entries.Put(k, httpSendCacheEntry{response: v})
}

func (cache *httpSendCache) putError(k ast.Value, v error) {
	cache.entries.Put(k, httpSendCacheEntry{error: v})
}

// In the BuiltinContext cache we only store a single entry that points to
// our ValueMap which is the "real" http.send() cache.
func getHTTPSendCache(bctx BuiltinContext) *httpSendCache {
	raw, ok := bctx.Cache.Get(httpSendBuiltinCacheKey)
	if !ok {
		// Initialize if it isn't there
		c := newHTTPSendCache()
		bctx.Cache.Put(httpSendBuiltinCacheKey, c)
		return c
	}

	c, ok := raw.(*httpSendCache)
	if !ok {
		return nil
	}
	return c
}

// checkHTTPSendCache checks for the given key's value in the cache
func checkHTTPSendCache(bctx BuiltinContext, key ast.Object) (ast.Value, error) {
	requestCache := getHTTPSendCache(bctx)
	if requestCache == nil {
		return nil, nil
	}

	v := requestCache.get(key)
	if v != nil {
		if v.error != nil {
			return nil, v.error
		}
		if v.response != nil {
			return *v.response, nil
		}
		// This should never happen
	}

	return nil, nil
}

func insertIntoHTTPSendCache(bctx BuiltinContext, key ast.Object, value ast.Value) {
	requestCache := getHTTPSendCache(bctx)
	if requestCache == nil {
		// Should never happen.. if it does just skip caching the value
		// FIXME: return error instead, to prevent inconsistencies?
		return
	}
	requestCache.putResponse(key, &value)
}

func insertErrorIntoHTTPSendCache(bctx BuiltinContext, key ast.Object, err error) {
	requestCache := getHTTPSendCache(bctx)
	if requestCache == nil {
		// Should never happen.. if it does just skip caching the value
		// FIXME: return error instead, to prevent inconsistencies?
		return
	}
	requestCache.putError(key, err)
}

// checkHTTPSendInterQueryCache checks for the given key's value in the inter-query cache
func (c *interQueryCache) checkHTTPSendInterQueryCache() (ast.Value, error) {
	requestCache := c.bctx.InterQueryBuiltinCache

	cachedValue, found := requestCache.Get(c.key)
	if !found {
		return nil, nil
	}

	value, cerr := requestCache.Clone(cachedValue)
	if cerr != nil {
		return nil, handleHTTPSendErr(c.bctx, cerr)
	}

	c.bctx.Metrics.Counter(httpSendInterQueryCacheHits).Incr()
	var cachedRespData *interQueryCacheData

	switch v := value.(type) {
	case *interQueryCacheValue:
		var err error
		cachedRespData, err = v.copyCacheData()
		if err != nil {
			return nil, err
		}
	case *interQueryCacheData:
		cachedRespData = v
	default:
		return nil, nil
	}

	if getCurrentTime(c.bctx).Before(cachedRespData.ExpiresAt) {
		return cachedRespData.formatToAST(c.forceJSONDecode, c.forceYAMLDecode)
	}

	var err error
	c.httpReq, c.httpClient, err = createHTTPRequest(c.bctx, c.key)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	headers := parseResponseHeaders(cachedRespData.Headers)

	// check with the server if the stale response is still up-to-date.
	// If server returns a new response (ie. status_code=200), update the cache with the new response
	// If server returns an unmodified response (ie. status_code=304), update the headers for the existing response
	result, modified, err := revalidateCachedResponse(c.httpReq, c.httpClient, c.key, headers)
	requestCache.Delete(c.key)
	if err != nil || result == nil {
		return nil, err
	}

	defer result.Body.Close()

	if !modified {
		// update the headers in the cached response with their corresponding values from the 304 (Not Modified) response
		for headerName, values := range result.Header {
			cachedRespData.Headers.Del(headerName)
			for _, v := range values {
				cachedRespData.Headers.Add(headerName, v)
			}
		}

		if forceCaching(c.forceCacheParams) {
			createdAt := getCurrentTime(c.bctx)
			cachedRespData.ExpiresAt = createdAt.Add(time.Second * time.Duration(c.forceCacheParams.forceCacheDurationSeconds))
		} else {
			expiresAt, err := expiryFromHeaders(result.Header)
			if err != nil {
				return nil, err
			}
			cachedRespData.ExpiresAt = expiresAt
		}

		cachingMode, err := getCachingMode(c.key)
		if err != nil {
			return nil, err
		}

		var pcv cache.InterQueryCacheValue

		if cachingMode == defaultCachingMode {
			pcv, err = cachedRespData.toCacheValue()
			if err != nil {
				return nil, err
			}
		} else {
			pcv = cachedRespData
		}

		c.bctx.InterQueryBuiltinCache.InsertWithExpiry(c.key, pcv, cachedRespData.ExpiresAt)

		return cachedRespData.formatToAST(c.forceJSONDecode, c.forceYAMLDecode)
	}

	newValue, respBody, err := formatHTTPResponseToAST(result, c.forceJSONDecode, c.forceYAMLDecode)
	if err != nil {
		return nil, err
	}

	if err := insertIntoHTTPSendInterQueryCache(c.bctx, c.key, result, respBody, c.forceCacheParams); err != nil {
		return nil, err
	}

	return newValue, nil
}

// insertIntoHTTPSendInterQueryCache inserts given key and value in the inter-query cache
func insertIntoHTTPSendInterQueryCache(bctx BuiltinContext, key ast.Value, resp *http.Response, respBody []byte, cacheParams *forceCacheParams) error {
	if resp == nil || (!forceCaching(cacheParams) && !canStore(resp.Header)) || !cacheableCodes.Contains(ast.IntNumberTerm(resp.StatusCode)) {
		return nil
	}

	requestCache := bctx.InterQueryBuiltinCache

	obj, ok := key.(ast.Object)
	if !ok {
		return fmt.Errorf("interface conversion error")
	}

	cachingMode, err := getCachingMode(obj)
	if err != nil {
		return err
	}

	var pcv cache.InterQueryCacheValue
	var pcvData *interQueryCacheData
	if cachingMode == defaultCachingMode {
		pcv, pcvData, err = newInterQueryCacheValue(bctx, resp, respBody, cacheParams)
	} else {
		pcvData, err = newInterQueryCacheData(bctx, resp, respBody, cacheParams)
		pcv = pcvData
	}

	if err != nil {
		return err
	}

	requestCache.InsertWithExpiry(key, pcv, pcvData.ExpiresAt)
	return nil
}

func createAllowedKeys() {
	for _, element := range allowedKeyNames {
		allowedKeys.Add(ast.StringTerm(element))
	}
}

func createCacheableHTTPStatusCodes() {
	for _, element := range cacheableHTTPStatusCodes {
		cacheableCodes.Add(ast.IntNumberTerm(element))
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

func getNumberValFromReqObj(req ast.Object, key *ast.Term) (int, error) {
	term := req.Get(key)
	if term == nil {
		return 0, nil
	}

	if t, ok := term.Value.(ast.Number); ok {
		num, ok := t.Int()
		if !ok || num < 0 {
			return 0, fmt.Errorf("invalid value %v for field %v", t.String(), key.String())
		}
		return num, nil
	}

	return 0, fmt.Errorf("invalid value %v for field %v", term.String(), key.String())
}

func getCachingMode(req ast.Object) (cachingMode, error) {
	key := ast.StringTerm("caching_mode")
	var s ast.String
	var ok bool
	if v := req.Get(key); v != nil {
		if s, ok = v.Value.(ast.String); !ok {
			return "", fmt.Errorf("invalid value for %v field", key.String())
		}

		switch cachingMode(s) {
		case defaultCachingMode, cachingModeDeserialized:
			return cachingMode(s), nil
		default:
			return "", fmt.Errorf("invalid value specified for %v field: %v", key.String(), string(s))
		}
	}
	return defaultCachingMode, nil
}

type interQueryCacheValue struct {
	Data []byte
}

func newInterQueryCacheValue(bctx BuiltinContext, resp *http.Response, respBody []byte, cacheParams *forceCacheParams) (*interQueryCacheValue, *interQueryCacheData, error) {
	data, err := newInterQueryCacheData(bctx, resp, respBody, cacheParams)
	if err != nil {
		return nil, nil, err
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	return &interQueryCacheValue{Data: b}, data, nil
}

func (cb interQueryCacheValue) Clone() (cache.InterQueryCacheValue, error) {
	dup := make([]byte, len(cb.Data))
	copy(dup, cb.Data)
	return &interQueryCacheValue{Data: dup}, nil
}

func (cb interQueryCacheValue) SizeInBytes() int64 {
	return int64(len(cb.Data))
}

func (cb *interQueryCacheValue) copyCacheData() (*interQueryCacheData, error) {
	var res interQueryCacheData
	err := util.UnmarshalJSON(cb.Data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type interQueryCacheData struct {
	RespBody   []byte
	Status     string
	StatusCode int
	Headers    http.Header
	ExpiresAt  time.Time
}

func forceCaching(cacheParams *forceCacheParams) bool {
	return cacheParams != nil && cacheParams.forceCacheDurationSeconds > 0
}

func expiryFromHeaders(headers http.Header) (time.Time, error) {
	var expiresAt time.Time
	maxAge, err := parseMaxAgeCacheDirective(parseCacheControlHeader(headers))
	if err != nil {
		return time.Time{}, err
	}
	if maxAge != -1 {
		createdAt, err := getResponseHeaderDate(headers)
		if err != nil {
			return time.Time{}, err
		}
		expiresAt = createdAt.Add(time.Second * time.Duration(maxAge))
	} else {
		expiresAt = getResponseHeaderExpires(headers)
	}
	return expiresAt, nil
}

func newInterQueryCacheData(bctx BuiltinContext, resp *http.Response, respBody []byte, cacheParams *forceCacheParams) (*interQueryCacheData, error) {
	var expiresAt time.Time

	if forceCaching(cacheParams) {
		createdAt := getCurrentTime(bctx)
		expiresAt = createdAt.Add(time.Second * time.Duration(cacheParams.forceCacheDurationSeconds))
	} else {
		var err error
		expiresAt, err = expiryFromHeaders(resp.Header)
		if err != nil {
			return nil, err
		}
	}

	cv := interQueryCacheData{
		ExpiresAt:  expiresAt,
		RespBody:   respBody,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header}

	return &cv, nil
}

func (c *interQueryCacheData) formatToAST(forceJSONDecode, forceYAMLDecode bool) (ast.Value, error) {
	return prepareASTResult(c.Headers, forceJSONDecode, forceYAMLDecode, c.RespBody, c.Status, c.StatusCode)
}

func (c *interQueryCacheData) toCacheValue() (*interQueryCacheValue, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return &interQueryCacheValue{Data: b}, nil
}

func (c *interQueryCacheData) SizeInBytes() int64 {
	return 0
}

func (c *interQueryCacheData) Clone() (cache.InterQueryCacheValue, error) {
	dup := make([]byte, len(c.RespBody))
	copy(dup, c.RespBody)

	return &interQueryCacheData{
		ExpiresAt:  c.ExpiresAt,
		RespBody:   dup,
		Status:     c.Status,
		StatusCode: c.StatusCode,
		Headers:    c.Headers.Clone()}, nil
}

type responseHeaders struct {
	etag         string // identifier for a specific version of the response
	lastModified string // date and time response was last modified as per origin server
}

// deltaSeconds specifies a non-negative integer, representing
// time in seconds: http://tools.ietf.org/html/rfc7234#section-1.2.1
type deltaSeconds int32

func parseResponseHeaders(headers http.Header) *responseHeaders {
	result := responseHeaders{}

	result.etag = headers.Get("etag")

	result.lastModified = headers.Get("last-modified")

	return &result
}

func revalidateCachedResponse(req *http.Request, client *http.Client, inputReqObj ast.Object, headers *responseHeaders) (*http.Response, bool, error) {
	etag := headers.etag
	lastModified := headers.lastModified

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

	response, err := executeHTTPRequest(cloneReq, client, inputReqObj)
	if err != nil {
		return nil, false, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return response, true, nil

	case http.StatusNotModified:
		return response, false, nil
	}
	util.Close(response)
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
	return http.ParseTime(dateHeader)
}

func getResponseHeaderExpires(headers http.Header) time.Time {
	expiresHeader := headers.Get("expires")
	if expiresHeader == "" {
		return time.Time{}
	}

	date, err := http.ParseTime(expiresHeader)
	if err != nil {
		// servers can set `Expires: 0` which is an invalid date to indicate expired content
		return time.Time{}
	}

	return date
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

func formatHTTPResponseToAST(resp *http.Response, forceJSONDecode, forceYAMLDecode bool) (ast.Value, []byte, error) {

	resultRawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	resultObj, err := prepareASTResult(resp.Header, forceJSONDecode, forceYAMLDecode, resultRawBody, resp.Status, resp.StatusCode)
	if err != nil {
		return nil, nil, err
	}

	return resultObj, resultRawBody, nil
}

func prepareASTResult(headers http.Header, forceJSONDecode, forceYAMLDecode bool, body []byte, status string, statusCode int) (ast.Value, error) {
	var resultBody interface{}

	// If the response body cannot be JSON/YAML decoded,
	// an error will not be returned. Instead, the "body" field
	// in the result will be null.
	switch {
	case forceJSONDecode || isContentType(headers, "application/json"):
		_ = util.UnmarshalJSON(body, &resultBody)
	case forceYAMLDecode || isContentType(headers, "application/yaml", "application/x-yaml"):
		_ = util.Unmarshal(body, &resultBody)
	}

	result := make(map[string]interface{})
	result["status"] = status
	result["status_code"] = statusCode
	result["body"] = resultBody
	result["raw_body"] = string(body)
	result["headers"] = getResponseHeaders(headers)

	resultObj, err := ast.InterfaceToValue(result)
	if err != nil {
		return nil, err
	}

	return resultObj, nil
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
	InsertErrorIntoCache(err error)
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
	forceYAMLDecode  bool
	forceCacheParams *forceCacheParams
}

func newInterQueryCache(bctx BuiltinContext, key ast.Object, forceCacheParams *forceCacheParams) (*interQueryCache, error) {
	return &interQueryCache{bctx: bctx, key: key, forceCacheParams: forceCacheParams}, nil
}

// CheckCache checks the cache for the value of the key set on this object
func (c *interQueryCache) CheckCache() (ast.Value, error) {
	var err error

	// Checking the intra-query cache first ensures consistency of errors and HTTP responses within a query.
	resp, err := checkHTTPSendCache(c.bctx, c.key)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}

	c.forceJSONDecode, err = getBoolValFromReqObj(c.key, ast.StringTerm("force_json_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}
	c.forceYAMLDecode, err = getBoolValFromReqObj(c.key, ast.StringTerm("force_yaml_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	resp, err = c.checkHTTPSendInterQueryCache()
	// Always insert the result of the inter-query cache into the intra-query cache, to maintain consistency within the same query.
	if err != nil {
		insertErrorIntoHTTPSendCache(c.bctx, c.key, err)
	}
	if resp != nil {
		insertIntoHTTPSendCache(c.bctx, c.key, resp)
	}
	return resp, err
}

// InsertIntoCache inserts the key set on this object into the cache with the given value
func (c *interQueryCache) InsertIntoCache(value *http.Response) (ast.Value, error) {
	result, respBody, err := formatHTTPResponseToAST(value, c.forceJSONDecode, c.forceYAMLDecode)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	// Always insert into the intra-query cache, to maintain consistency within the same query.
	insertIntoHTTPSendCache(c.bctx, c.key, result)

	// We ignore errors when populating the inter-query cache, because we've already populated the intra-cache,
	// and query consistency is our primary concern.
	_ = insertIntoHTTPSendInterQueryCache(c.bctx, c.key, value, respBody, c.forceCacheParams)
	return result, nil
}

func (c *interQueryCache) InsertErrorIntoCache(err error) {
	insertErrorIntoHTTPSendCache(c.bctx, c.key, err)
}

// ExecuteHTTPRequest executes a HTTP request
func (c *interQueryCache) ExecuteHTTPRequest() (*http.Response, error) {
	var err error
	c.httpReq, c.httpClient, err = createHTTPRequest(c.bctx, c.key)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	return executeHTTPRequest(c.httpReq, c.httpClient, c.key)
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
	return checkHTTPSendCache(c.bctx, c.key)
}

// InsertIntoCache inserts the key set on this object into the cache with the given value
func (c *intraQueryCache) InsertIntoCache(value *http.Response) (ast.Value, error) {
	forceJSONDecode, err := getBoolValFromReqObj(c.key, ast.StringTerm("force_json_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}
	forceYAMLDecode, err := getBoolValFromReqObj(c.key, ast.StringTerm("force_yaml_decode"))
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	result, _, err := formatHTTPResponseToAST(value, forceJSONDecode, forceYAMLDecode)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}

	if cacheableCodes.Contains(ast.IntNumberTerm(value.StatusCode)) {
		insertIntoHTTPSendCache(c.bctx, c.key, result)
	}

	return result, nil
}

func (c *intraQueryCache) InsertErrorIntoCache(err error) {
	insertErrorIntoHTTPSendCache(c.bctx, c.key, err)
}

// ExecuteHTTPRequest executes a HTTP request
func (c *intraQueryCache) ExecuteHTTPRequest() (*http.Response, error) {
	httpReq, httpClient, err := createHTTPRequest(c.bctx, c.key)
	if err != nil {
		return nil, handleHTTPSendErr(c.bctx, err)
	}
	return executeHTTPRequest(httpReq, httpClient, c.key)
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
