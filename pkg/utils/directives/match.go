/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package directives

import (
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"
)

// Following code copied from github.com/caddyserver/caddy/modules/caddyhttp/matchers.go

type (
	MatchPath []string
)

func (m MatchPath) Match(req *http.Request) bool {
	// Even though RFC 9110 says that path matching is case-sensitive
	// (https://www.rfc-editor.org/rfc/rfc9110.html#section-4.2.3),
	// we do case-insensitive matching to mitigate security issues
	// related to differences between operating systems, applications,
	// etc; if case-sensitive matching is needed, the regex matcher
	// can be used instead.
	reqPath := strings.ToLower(req.URL.Path)

	// See #2917; Windows ignores trailing dots and spaces
	// when accessing files (sigh), potentially causing a
	// security risk (cry) if PHP files end up being served
	// as static files, exposing the source code, instead of
	// being matched by *.php to be treated as PHP scripts.
	if runtime.GOOS == "windows" { // issue #5613
		reqPath = strings.TrimRight(reqPath, ". ")
	}

	repl := req.Context().Value(ReplacerCtxKey).(*Replacer)

	for _, matchPattern := range m {
		matchPattern = repl.ReplaceAll(matchPattern, "")

		// special case: whole path is wildcard; this is unnecessary
		// as it matches all requests, which is the same as no matcher
		if matchPattern == "*" {
			return true
		}

		// Clean the path, merge doubled slashes, etc.
		// This ensures maliciously crafted requests can't bypass
		// the path matcher. See #4407. Good security posture
		// requires that we should do all we can to reduce any
		// funny-looking paths into "normalized" forms such that
		// weird variants can't sneak by.
		//
		// How we clean the path depends on the kind of pattern:
		// we either merge slashes or we don't. If the pattern
		// has double slashes, we preserve them in the path.
		//
		// TODO: Despite the fact that the *vast* majority of path
		// matchers have only 1 pattern, a possible optimization is
		// to remember the cleaned form of the path for future
		// iterations; it's just that the way we clean depends on
		// the kind of pattern.

		mergeSlashes := !strings.Contains(matchPattern, "//")

		// if '%' appears in the match pattern, we interpret that to mean
		// the intent is to compare that part of the path in raw/escaped
		// space; i.e. "%40"=="%40", not "@", and "%2F"=="%2F", not "/"
		if strings.Contains(matchPattern, "%") {
			reqPathForPattern := CleanPath(req.URL.EscapedPath(), mergeSlashes)
			if m.matchPatternWithEscapeSequence(reqPathForPattern, matchPattern) {
				return true
			}

			// doing prefix/suffix/substring matches doesn't make sense
			continue
		}

		reqPathForPattern := CleanPath(reqPath, mergeSlashes)

		// for substring, prefix, and suffix matching, only perform those
		// special, fast matches if they are the only wildcards in the pattern;
		// otherwise we assume a globular match if any * appears in the middle

		// special case: first and last characters are wildcard,
		// treat it as a fast substring match
		if strings.Count(matchPattern, "*") == 2 &&
			strings.HasPrefix(matchPattern, "*") &&
			strings.HasSuffix(matchPattern, "*") &&
			strings.Count(matchPattern, "*") == 2 {
			if strings.Contains(reqPathForPattern, matchPattern[1:len(matchPattern)-1]) {
				return true
			}
			continue
		}

		// only perform prefix/suffix match if it is the only wildcard...
		// I think that is more correct most of the time
		if strings.Count(matchPattern, "*") == 1 {
			// special case: first character is a wildcard,
			// treat it as a fast suffix match
			if strings.HasPrefix(matchPattern, "*") {
				if strings.HasSuffix(reqPathForPattern, matchPattern[1:]) {
					return true
				}
				continue
			}

			// special case: last character is a wildcard,
			// treat it as a fast prefix match
			if strings.HasSuffix(matchPattern, "*") {
				if strings.HasPrefix(reqPathForPattern, matchPattern[:len(matchPattern)-1]) {
					return true
				}
				continue
			}
		}

		// at last, use globular matching, which also is exact matching
		// if there are no glob/wildcard chars; we ignore the error here
		// because we can't handle it anyway
		matches, _ := path.Match(matchPattern, reqPathForPattern)
		if matches {
			return true
		}
	}
	return false
}

func (MatchPath) matchPatternWithEscapeSequence(escapedPath, matchPath string) bool {
	// We would just compare the pattern against r.URL.Path,
	// but the pattern contains %, indicating that we should
	// compare at least some part of the path in raw/escaped
	// space, not normalized space; so we build the string we
	// will compare against by adding the normalized parts
	// of the path, then switching to the escaped parts where
	// the pattern hints to us wherever % is present.
	var sb strings.Builder

	// iterate the pattern and escaped path in lock-step;
	// increment iPattern every time we consume a char from the pattern,
	// increment iPath every time we consume a char from the path;
	// iPattern and iPath are our cursors/iterator positions for each string
	var iPattern, iPath int
	for {
		if iPattern >= len(matchPath) || iPath >= len(escapedPath) {
			break
		}

		// get the next character from the request path

		pathCh := string(escapedPath[iPath])
		var escapedPathCh string

		// normalize (decode) escape sequences
		if pathCh == "%" && len(escapedPath) >= iPath+3 {
			// hold onto this in case we find out the intent is to match in escaped space here;
			// we lowercase it even though technically the spec says: "For consistency, URI
			// producers and normalizers should use uppercase hexadecimal digits for all percent-
			// encodings" (RFC 3986 section 2.1) - we lowercased the matcher pattern earlier in
			// provisioning so we do the same here to gain case-insensitivity in equivalence;
			// besides, this string is never shown visibly
			escapedPathCh = strings.ToLower(escapedPath[iPath : iPath+3])

			var err error
			pathCh, err = url.PathUnescape(escapedPathCh)
			if err != nil {
				// should be impossible unless EscapedPath() is giving us an invalid sequence!
				return false
			}
			iPath += 2 // escape sequence is 2 bytes longer than normal char
		}

		// now get the next character from the pattern

		normalize := true
		switch matchPath[iPattern] {
		case '%':
			// escape sequence

			// if not a wildcard ("%*"), compare literally; consume next two bytes of pattern
			if len(matchPath) >= iPattern+3 && matchPath[iPattern+1] != '*' {
				sb.WriteString(escapedPathCh)
				iPath++
				iPattern += 2
				break
			}

			// escaped wildcard sequence; consume next byte only ('*')
			iPattern++
			normalize = false

			fallthrough
		case '*':
			// wildcard, so consume until next matching character
			remaining := escapedPath[iPath:]
			until := len(escapedPath) - iPath // go until end of string...
			if iPattern < len(matchPath)-1 {  // ...unless the * is not at the end
				nextCh := matchPath[iPattern+1]
				until = strings.IndexByte(remaining, nextCh)
				if until == -1 {
					// terminating char of wildcard span not found, so definitely no match
					return false
				}
			}
			if until == 0 {
				// empty span; nothing to add on this iteration
				break
			}
			next := remaining[:until]
			if normalize {
				var err error
				next, err = url.PathUnescape(next)
				if err != nil {
					return false // should be impossible anyway
				}
			}
			sb.WriteString(next)
			iPath += until
		default:
			sb.WriteString(pathCh)
			iPath++
		}

		iPattern++
	}

	// we can now treat rawpath globs (%*) as regular globs (*)
	matchPath = strings.ReplaceAll(matchPath, "%*", "*")

	// ignore error here because we can't handle it anyway=
	matches, _ := path.Match(matchPath, sb.String())
	return matches
}
