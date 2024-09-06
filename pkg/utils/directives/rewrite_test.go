/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package directives

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestRewrite(t *testing.T) {
	repl := NewReplacer()

	for i, tc := range []struct {
		input, expect *http.Request
		rule          Rewrite
	}{
		{
			rule:   Rewrite{StripPathPrefix: "/api"},
			input:  newRequest(t, "GET", "/api"),
			expect: newRequest(t, "GET", "/"),
		},
		{
			rule:   Rewrite{StripPathSuffix: ".html"},
			input:  newRequest(t, "GET", "/index.html"),
			expect: newRequest(t, "GET", "/index"),
		},
		{
			rule: Rewrite{URISubstring: []substrReplacer{
				{
					Find:    "/docs/",
					Replace: "/v1/docs/",
					Limit:   0,
				},
			}},
			input:  newRequest(t, "GET", "/docs/"),
			expect: newRequest(t, "GET", "/v1/docs/"),
		},
		{
			rule: Rewrite{PathRegexp: []*regexReplacer{
				{
					Find:    "/{2,}",
					Replace: "/",
				},
			}},
			input:  newRequest(t, "GET", "/doc//readme.md"),
			expect: newRequest(t, "GET", "/doc/readme.md"),
		},
	} {
		// copy the original input just enough so that we can
		// compare it after the rewrite to see if it changed
		urlCopy := *tc.input.URL
		originalInput := &http.Request{
			Method:     tc.input.Method,
			RequestURI: tc.input.RequestURI,
			URL:        &urlCopy,
		}

		// populate the replacer just enough for our tests
		repl.Set("http.request.uri", tc.input.RequestURI)
		repl.Set("http.request.uri.path", tc.input.URL.Path)
		repl.Set("http.request.uri.query", tc.input.URL.RawQuery)

		for _, rep := range tc.rule.PathRegexp {
			re, err := regexp.Compile(rep.Find)
			if err != nil {
				t.Fatal(err)
			}
			rep.re = re
		}

		changed := tc.rule.Rewrite(tc.input, repl)
		if expected, actual := !reqEqual(originalInput, tc.input), changed; expected != actual {
			t.Errorf("Test %d: Expected changed=%t but was %t", i, expected, actual)
		}
		if tc.rule.StripPathPrefix != "" {
			t.Logf("Test UriRule \"uri strip_prefix %v\" ==> rewrite \"%v\" to \"%v\"", tc.rule.StripPathPrefix, originalInput.URL, tc.input.URL)
		} else if tc.rule.StripPathSuffix != "" {
			t.Logf("Test UriRule \"uri strip_suffix %v\" ==> rewrite \"%v\" to \"%v\"", tc.rule.StripPathSuffix, originalInput.URL, tc.input.URL)
		} else if tc.rule.URISubstring != nil {
			t.Logf("Test UriRule \"uri replace %s %s\" ==> rewrite \"%v\" to \"%v\"", tc.rule.URISubstring[0].Find, tc.rule.URISubstring[0].Replace, originalInput.URL, tc.input.URL)
		} else if tc.rule.PathRegexp != nil {
			t.Logf("Test UriRule \"uri path_regexp %s %s\" ==> rewrite \"%v\" to \"%v\"", (*tc.rule.PathRegexp[0]).Find, (*tc.rule.PathRegexp[0]).Replace, originalInput.URL, tc.input.URL)
		}

	}
}

func TestPathRewriteRule(t *testing.T) {
	for i, tc := range []struct {
		rr           []RewriteRule // not URI-encoded because not parsing from a URI
		input        string        // should be valid URI encoding (escaped) since it will become part of a request
		expect       bool
		provisionErr bool
	}{
		{
			rr: NewRewriteRulesWithOptions([]string{
				"* /foo.html",
			}, WithRewriteFilter),
			input:  "/",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"/api/* ?a=b",
			}, WithRewriteFilter),
			input:  "/api/abc",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"/api/* ?{query}&a=b",
			}, WithRewriteFilter),
			input:  "/api/abc",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"* /index.php?{query}&p={path}",
			}, WithRewriteFilter),
			input:  "/foo/bar",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"/api",
			}, WithStripPrefixFilter),
			input:  "/api/v1",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				".html",
			}, WithStripSuffixFilter),
			input:  "/index.html",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"/docs/ /v1/docs/",
			}, WithReplaceFilter),
			input:  "/docs/go",
			expect: true,
		},
		{
			rr: NewRewriteRulesWithOptions([]string{
				"/{2,} /",
			}, WithPathRegexpFilter),
			input:  "/doc//readme.md",
			expect: true,
		},
	} {
		u, err := url.ParseRequestURI(tc.input)
		if err != nil {
			t.Fatalf("Test %d (%v): Invalid request URI (should be rejected by Go's HTTP server): %v", i, tc.input, err)
		}
		req := &http.Request{URL: u}
		repl := NewReplacer()
		repl.Set("query", req.URL.RawQuery)
		repl.Set("path", req.URL.Path)
		//t.Logf("Init ENV with: {\"query\":\"%v\", \"path\": \"%v\"}", req.URL.RawQuery, req.URL.Path)
		ctx := context.WithValue(req.Context(), ReplacerCtxKey, repl)
		req = req.WithContext(ctx)

		for _, r := range tc.rr {
			oldRUL := req.URL.Path
			actual, err := r.Exec(req)
			if err != nil {
				t.Errorf("Test RewriteRule \"rewrite %v %v\" ==> Err %v", r.Match[0], r.Rewrite.URI, err)
				continue
			}
			if actual != tc.expect {
				t.Errorf("Test RewriteRule \"rewrite %v %v\" ==> Expected %t, got %t for '%s'", r.Match[0], r.Rewrite.URI, tc.expect, actual, tc.input)
				continue
			}

			if r.Rewrite.StripPathPrefix != "" {
				t.Logf("Test RewriteRule \"strip_prefix %v\" ==> rewrite \"%v\" to \"%v\"", r.Rewrite.StripPathPrefix, oldRUL, req.URL)
			} else if r.Rewrite.StripPathSuffix != "" {
				t.Logf("Test RewriteRule \"strip_suffix %v\" ==> rewrite \"%v\" to \"%v\"", r.Rewrite.StripPathSuffix, oldRUL, req.URL)
			} else if r.Rewrite.URISubstring != nil {
				t.Logf("Test RewriteRule \"replace %s %s\" ==> rewrite \"%v\" to \"%v\"", r.Rewrite.URISubstring[0].Find, r.Rewrite.URISubstring[0].Replace, oldRUL, req.URL)
			} else if r.Rewrite.PathRegexp != nil {
				t.Logf("Test RewriteRule \"path_regexp %s %s\" ==> rewrite \"%v\" to \"%v\"", (*r.Rewrite.PathRegexp[0]).Find, (*r.Rewrite.PathRegexp[0]).Replace, oldRUL, req.URL)
			} else if r.Rewrite.URI != "" {
				t.Logf("Test RewriteRule \"rewrite %s %s\" ==> rewrite \"%v\" to \"%v\"", r.Match[0], r.Rewrite.URI, oldRUL, req.URL)
			}
		}
	}
}

func newRequest(t *testing.T, method, uri string) *http.Request {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	req.RequestURI = req.URL.RequestURI() // simulate incoming request
	return req
}

func reqEqual(r1, r2 *http.Request) bool {
	if r1.Method != r2.Method {
		return false
	}
	if r1.RequestURI != r2.RequestURI {
		return false
	}
	if (r1.URL == nil && r2.URL != nil) || (r1.URL != nil && r2.URL == nil) {
		return false
	}
	if r1.URL == nil && r2.URL == nil {
		return true
	}
	return r1.URL.Scheme == r2.URL.Scheme &&
		r1.URL.Host == r2.URL.Host &&
		r1.URL.Path == r2.URL.Path &&
		r1.URL.RawPath == r2.URL.RawPath &&
		r1.URL.RawQuery == r2.URL.RawQuery &&
		r1.URL.Fragment == r2.URL.Fragment
}
