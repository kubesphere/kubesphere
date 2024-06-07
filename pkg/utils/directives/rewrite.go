/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package directives

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Following code copied from github.com/caddyserver/caddy/modules/caddyhttp/rewrite/rewrite.go

type Rewrite struct {
	Method          string           `json:"method,omitempty"`
	URI             string           `json:"uri,omitempty"`
	StripPathPrefix string           `json:"strip_path_prefix,omitempty"`
	StripPathSuffix string           `json:"strip_path_suffix,omitempty"`
	URISubstring    []substrReplacer `json:"uri_substring,omitempty"`
	PathRegexp      []*regexReplacer `json:"path_regexp,omitempty"`
}

func (r *Rewrite) Rewrite(req *http.Request, repl *Replacer) bool {
	oldMethod := req.Method
	oldURI := req.RequestURI

	// method
	if r.Method != "" {
		req.Method = strings.ToUpper(repl.ReplaceAll(r.Method, ""))
	}

	// uri (path, query string and... fragment, because why not)
	if uri := r.URI; uri != "" {
		// find the bounds of each part of the URI that exist
		pathStart, qsStart, fragStart := -1, -1, -1
		pathEnd, qsEnd := -1, -1
	loop:
		for i, ch := range uri {
			switch {
			case ch == '?' && qsStart < 0:
				pathEnd, qsStart = i, i+1
			case ch == '#' && fragStart < 0: // everything after fragment is fragment (very clear in RFC 3986 section 4.2)
				if qsStart < 0 {
					pathEnd = i
				} else {
					qsEnd = i
				}
				fragStart = i + 1
				break loop
			case pathStart < 0 && qsStart < 0:
				pathStart = i
			}
		}
		if pathStart >= 0 && pathEnd < 0 {
			pathEnd = len(uri)
		}
		if qsStart >= 0 && qsEnd < 0 {
			qsEnd = len(uri)
		}

		// isolate the three main components of the URI
		var path, query, frag string
		if pathStart > -1 {
			path = uri[pathStart:pathEnd]
		}
		if qsStart > -1 {
			query = uri[qsStart:qsEnd]
		}
		if fragStart > -1 {
			frag = uri[fragStart:]
		}

		// build components which are specified, and store them
		// in a temporary variable so that they all read the
		// same version of the URI
		var newPath, newQuery, newFrag string

		if path != "" {
			// replace the `path` placeholder to escaped path
			pathPlaceholder := "{http.request.uri.path}"
			if strings.Contains(path, pathPlaceholder) {
				path = strings.ReplaceAll(path, pathPlaceholder, req.URL.EscapedPath())
			}

			newPath = repl.ReplaceAll(path, "")
		}

		// before continuing, we need to check if a query string
		// snuck into the path component during replacements
		if before, after, found := strings.Cut(newPath, "?"); found {
			// recompute; new path contains a query string
			var injectedQuery string
			newPath, injectedQuery = before, after
			// don't overwrite explicitly-configured query string
			if query == "" {
				query = injectedQuery
			}
		}

		if query != "" {
			newQuery = buildQueryString(query, repl)
		}
		if frag != "" {
			newFrag = repl.ReplaceAll(frag, "")
		}

		// update the URI with the new components
		// only after building them
		if pathStart >= 0 {
			if path, err := url.PathUnescape(newPath); err != nil {
				req.URL.Path = newPath
			} else {
				req.URL.Path = path
			}
		}
		if qsStart >= 0 {
			req.URL.RawQuery = newQuery
		}
		if fragStart >= 0 {
			req.URL.Fragment = newFrag
		}
	}

	// strip path prefix or suffix
	if r.StripPathPrefix != "" {
		prefix := repl.ReplaceAll(r.StripPathPrefix, "")
		mergeSlashes := !strings.Contains(prefix, "//")
		changePath(req, func(escapedPath string) string {
			escapedPath = CleanPath(escapedPath, mergeSlashes)
			return trimPathPrefix(escapedPath, prefix)
		})
	}
	if r.StripPathSuffix != "" {
		suffix := repl.ReplaceAll(r.StripPathSuffix, "")
		mergeSlashes := !strings.Contains(suffix, "//")
		changePath(req, func(escapedPath string) string {
			escapedPath = CleanPath(escapedPath, mergeSlashes)
			return reverse(trimPathPrefix(reverse(escapedPath), reverse(suffix)))
		})
	}

	// substring replacements in URI
	for _, rep := range r.URISubstring {
		rep.do(req, repl)
	}

	// regular expression replacements on the path
	for _, rep := range r.PathRegexp {
		rep.do(req, repl)
	}

	// update the encoded copy of the URI
	req.RequestURI = req.URL.RequestURI()

	// return true if anything changed
	return req.Method != oldMethod || req.RequestURI != oldURI
}

func buildQueryString(qs string, repl *Replacer) string {
	var sb strings.Builder

	// first component must be key, which is the same
	// as if we just wrote a value in previous iteration
	wroteVal := true

	for len(qs) > 0 {
		// determine the end of this component, which will be at
		// the next equal sign or ampersand, whichever comes first
		nextEq, nextAmp := strings.Index(qs, "="), strings.Index(qs, "&")
		ampIsNext := nextAmp >= 0 && (nextAmp < nextEq || nextEq < 0)
		end := len(qs) // assume no delimiter remains...
		if ampIsNext {
			end = nextAmp // ...unless ampersand is first...
		} else if nextEq >= 0 && (nextEq < nextAmp || nextAmp < 0) {
			end = nextEq // ...or unless equal is first.
		}

		// consume the component and write the result
		comp := qs[:end]
		comp, _ = repl.ReplaceFunc(comp, func(name string, val any) (any, error) {
			if name == "http.request.uri.query" && wroteVal {
				return val, nil // already escaped
			}
			var valStr string
			switch v := val.(type) {
			case string:
				valStr = v
			case fmt.Stringer:
				valStr = v.String()
			case int:
				valStr = strconv.Itoa(v)
			default:
				valStr = fmt.Sprintf("%+v", v)
			}
			return url.QueryEscape(valStr), nil
		})
		if end < len(qs) {
			end++ // consume delimiter
		}
		qs = qs[end:]

		// if previous iteration wrote a value,
		// that means we are writing a key
		if wroteVal {
			if sb.Len() > 0 && len(comp) > 0 {
				sb.WriteRune('&')
			}
		} else {
			sb.WriteRune('=')
		}
		sb.WriteString(comp)

		// remember for the next iteration that we just wrote a value,
		// which means the next iteration MUST write a key
		wroteVal = ampIsNext
	}

	return sb.String()
}

func trimPathPrefix(escapedPath, prefix string) string {
	var iPath, iPrefix int
	for {
		if iPath >= len(escapedPath) || iPrefix >= len(prefix) {
			break
		}

		prefixCh := prefix[iPrefix]
		ch := string(escapedPath[iPath])

		if ch == "%" && prefixCh != '%' && len(escapedPath) >= iPath+3 {
			var err error
			ch, err = url.PathUnescape(escapedPath[iPath : iPath+3])
			if err != nil {
				// should be impossible unless EscapedPath() is returning invalid values!
				return escapedPath
			}
			iPath += 2
		}

		// prefix comparisons are case-insensitive to consistency with
		// path matcher, which is case-insensitive for good reasons
		if !strings.EqualFold(ch, string(prefixCh)) {
			return escapedPath
		}

		iPath++
		iPrefix++
	}

	// if we iterated through the entire prefix, we found it, so trim it
	if iPath >= len(prefix) {
		return escapedPath[iPath:]
	}

	// otherwise we did not find the prefix
	return escapedPath
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func changePath(req *http.Request, newVal func(pathOrRawPath string) string) {
	req.URL.RawPath = newVal(req.URL.EscapedPath())
	if p, err := url.PathUnescape(req.URL.RawPath); err == nil && p != "" {
		req.URL.Path = p
	} else {
		req.URL.Path = newVal(req.URL.Path)
	}
	// RawPath is only set if it's different from the normalized Path (std lib)
	if req.URL.RawPath == req.URL.Path {
		req.URL.RawPath = ""
	}
}

func CleanPath(p string, collapseSlashes bool) string {
	if collapseSlashes {
		return cleanPath(p)
	}

	// insert an invalid/impossible URI character into each two consecutive
	// slashes to expand empty path segments; then clean the path as usual,
	// and then remove the remaining temporary characters.
	const tmpCh = 0xff
	var sb strings.Builder
	for i, ch := range p {
		if ch == '/' && i > 0 && p[i-1] == '/' {
			sb.WriteByte(tmpCh)
		}
		sb.WriteRune(ch)
	}
	halfCleaned := cleanPath(sb.String())
	halfCleaned = strings.ReplaceAll(halfCleaned, string([]byte{tmpCh}), "")

	return halfCleaned
}

// cleanPath does path.Clean(p) but preserves any trailing slash.
func cleanPath(p string) string {
	cleaned := path.Clean(p)
	if cleaned != "/" && strings.HasSuffix(p, "/") {
		cleaned = cleaned + "/"
	}
	return cleaned
}

type RewriteRule struct {
	Match   MatchPath
	Rewrite Rewrite
}

func (rr *RewriteRule) Exec(req *http.Request) (change bool, err error) {
	var repl *Replacer

	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("RewriteRule Err %v", panicErr)
		}
	}()
	replCtx := req.Context().Value(ReplacerCtxKey)
	if replCtx == nil || replCtx.(*Replacer) == nil {
		repl := NewReplacer()
		repl.Set("query", req.URL.RawQuery)
		repl.Set("path", req.URL.Path)
		ctx := context.WithValue(req.Context(), ReplacerCtxKey, repl)
		req = req.WithContext(ctx)
	} else {
		repl = replCtx.(*Replacer)
	}
	if rr.Match == nil || rr.Match.Match(req) {
		return rr.Rewrite.Rewrite(req, repl), nil
	}
	return
}

type DirectiveFilter func(rr *[]RewriteRule, expr []string, exprLen int)
type WithDirectiveFilter func(wf *[]DirectiveFilter)

func NewRewriteRulesWithOptions(rules []string, directiveFilters ...WithDirectiveFilter) []RewriteRule {
	var rewriteRules = make([]RewriteRule, 0, 1)
	if rules == nil {
		return rewriteRules
	}
	var filter = make([]DirectiveFilter, 0, 1)
	// inject directiveFilter for filter rewrite/replace/path_regexp
	for _, directiveFilter := range directiveFilters {
		directiveFilter(&filter)
	}

	for _, rule := range rules {
		expr := strings.Split(rule, " ")
		exprLen := len(expr)
		for _, directiveFilter := range filter {
			directiveFilter(&rewriteRules, expr, exprLen)
		}
	}
	return rewriteRules
}

func WithRewriteFilter(df *[]DirectiveFilter) {
	filterFn := func(rr *[]RewriteRule, expr []string, exprLen int) {
		if exprLen >= 2 {
			rewrite := Rewrite{
				URI: expr[1],
			}
			matchRule := []string{
				expr[0],
			}
			*rr = append(*rr, RewriteRule{
				matchRule,
				rewrite,
			})
		}
	}
	*df = append(*df, filterFn)
}

func WithReplaceFilter(df *[]DirectiveFilter) {
	filterFn := func(rr *[]RewriteRule, expr []string, exprLen int) {
		if exprLen >= 2 {
			rewrite := Rewrite{
				URISubstring: []substrReplacer{
					{
						Find:    expr[0],
						Replace: expr[1],
						Limit:   0,
					},
				},
			}
			*rr = append(*rr, RewriteRule{
				nil,
				rewrite,
			})
		}
	}
	*df = append(*df, filterFn)
}

func WithPathRegexpFilter(df *[]DirectiveFilter) {
	filterFn := func(rr *[]RewriteRule, expr []string, exprLen int) {
		if exprLen >= 2 {
			re, err := regexp.Compile(expr[0])
			if err != nil {
				return
			}
			rewrite := Rewrite{
				PathRegexp: []*regexReplacer{
					{
						Find:    expr[0],
						Replace: expr[1],
						re:      re,
					},
				},
			}
			*rr = append(*rr, RewriteRule{
				nil,
				rewrite,
			})
		}
	}
	*df = append(*df, filterFn)
}

func WithStripPrefixFilter(df *[]DirectiveFilter) {
	filterFn := func(rr *[]RewriteRule, expr []string, exprLen int) {
		if exprLen >= 1 {
			*rr = append(*rr, RewriteRule{
				nil,
				Rewrite{
					StripPathPrefix: expr[0],
				},
			})
		}
	}
	*df = append(*df, filterFn)
}

func WithStripSuffixFilter(df *[]DirectiveFilter) {
	filterFn := func(rr *[]RewriteRule, expr []string, exprLen int) {
		if exprLen >= 1 {
			*rr = append(*rr, RewriteRule{
				nil,
				Rewrite{
					StripPathSuffix: expr[0],
				},
			})
		}
	}
	*df = append(*df, filterFn)
}

func HandlerRequest(req *http.Request, rules []string, directiveFilters ...WithDirectiveFilter) error {
	rewriteRules := NewRewriteRulesWithOptions(rules, directiveFilters...)
	for _, rewriteRule := range rewriteRules {
		if _, err := rewriteRule.Exec(req); err != nil {
			return err
		}
	}
	return nil
}
