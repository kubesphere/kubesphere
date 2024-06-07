/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package directives

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestPathMatcher(t *testing.T) {
	for i, tc := range []struct {
		match        MatchPath // not URI-encoded because not parsing from a URI
		input        string    // should be valid URI encoding (escaped) since it will become part of a request
		expect       bool
		provisionErr bool
	}{
		{
			match:  MatchPath{},
			input:  "/",
			expect: false,
		},
		{
			match:  MatchPath{"/"},
			input:  "/",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/bar"},
			input:  "/",
			expect: false,
		},
		{
			match:  MatchPath{"/foo/bar"},
			input:  "/foo/bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/bar/"},
			input:  "/foo/bar",
			expect: false,
		},
		{
			match:  MatchPath{"/foo/bar/"},
			input:  "/foo/bar/",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/bar/", "/other"},
			input:  "/other/",
			expect: false,
		},
		{
			match:  MatchPath{"/foo/bar/", "/other"},
			input:  "/other",
			expect: true,
		},
		{
			match:  MatchPath{"*.ext"},
			input:  "/foo/bar.ext",
			expect: true,
		},
		{
			match:  MatchPath{"*.php"},
			input:  "/index.PHP",
			expect: true,
		},
		{
			match:  MatchPath{"*.ext"},
			input:  "/foo/bar.ext",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/*/baz"},
			input:  "/foo/bar/baz",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/*/baz/bam"},
			input:  "/foo/bar/bam",
			expect: false,
		},
		{
			match:  MatchPath{"*substring*"},
			input:  "/foo/substring/bar.txt",
			expect: true,
		},
		{
			match:  MatchPath{"/foo"},
			input:  "/foo/bar",
			expect: false,
		},
		{
			match:  MatchPath{"/foo"},
			input:  "/foo/bar",
			expect: false,
		},
		{
			match:  MatchPath{"/foo"},
			input:  "/FOO",
			expect: true,
		},
		{
			match:  MatchPath{"/foo*"},
			input:  "/FOOOO",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/bar.txt"},
			input:  "/foo/BAR.txt",
			expect: true,
		},
		{
			match:  MatchPath{"/foo*"},
			input:  "//foo/bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo"},
			input:  "//foo",
			expect: true,
		},
		{
			match:  MatchPath{"//foo"},
			input:  "/foo",
			expect: false,
		},
		{
			match:  MatchPath{"//foo"},
			input:  "//foo",
			expect: true,
		},
		{
			match:  MatchPath{"/foo//*"},
			input:  "/foo//bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo//*"},
			input:  "/foo/%2Fbar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/%2F*"},
			input:  "/foo//bar",
			expect: false,
		},
		{
			match:  MatchPath{"/foo//bar"},
			input:  "/foo//bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/*//bar"},
			input:  "/foo///bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/%*//bar"},
			input:  "/foo///bar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo/%*//bar"},
			input:  "/foo//%2Fbar",
			expect: true,
		},
		{
			match:  MatchPath{"/foo*"},
			input:  "/%2F/foo",
			expect: true,
		},
		{
			match:  MatchPath{"*"},
			input:  "/",
			expect: true,
		},
		{
			match:  MatchPath{"*"},
			input:  "/foo/bar",
			expect: true,
		},
		{
			match:  MatchPath{"**"},
			input:  "/",
			expect: true,
		},
		{
			match:  MatchPath{"**"},
			input:  "/foo/bar",
			expect: true,
		},
		// notice these next three test cases are the same normalized path but are written differently
		{
			match:  MatchPath{"/%25@.txt"},
			input:  "/%25@.txt",
			expect: true,
		},
		{
			match:  MatchPath{"/%25@.txt"},
			input:  "/%25%40.txt",
			expect: true,
		},
		{
			match:  MatchPath{"/%25%40.txt"},
			input:  "/%25%40.txt",
			expect: true,
		},
		{
			match:  MatchPath{"/bands/*/*"},
			input:  "/bands/AC%2FDC/T.N.T",
			expect: false, // because * operates in normalized space
		},
		{
			match:  MatchPath{"/bands/%*/%*"},
			input:  "/bands/AC%2FDC/T.N.T",
			expect: true,
		},
		{
			match:  MatchPath{"/bands/%*/%*"},
			input:  "/bands/AC/DC/T.N.T",
			expect: false,
		},
		{
			match:  MatchPath{"/bands/%*"},
			input:  "/bands/AC/DC",
			expect: false, // not a suffix match
		},
		{
			match:  MatchPath{"/bands/%*"},
			input:  "/bands/AC%2FDC",
			expect: true,
		},
		{
			match:  MatchPath{"/foo%2fbar/baz"},
			input:  "/foo%2Fbar/baz",
			expect: true,
		},
		{
			match:  MatchPath{"/foo%2fbar/baz"},
			input:  "/foo/bar/baz",
			expect: false,
		},
		{
			match:  MatchPath{"/foo/bar/baz"},
			input:  "/foo%2fbar/baz",
			expect: true,
		},
	} {
		u, err := url.ParseRequestURI(tc.input)
		if err != nil {
			t.Fatalf("Test %d (%v): Invalid request URI (should be rejected by Go's HTTP server): %v", i, tc.input, err)
		}
		req := &http.Request{URL: u}
		repl := NewReplacer()
		ctx := context.WithValue(req.Context(), ReplacerCtxKey, repl)
		req = req.WithContext(ctx)

		actual := tc.match.Match(req)
		if actual != tc.expect {
			t.Errorf("Test %d %v: Expected %t, got %t for '%s'", i, tc.match, tc.expect, actual, tc.input)
			continue
		}
	}
}
