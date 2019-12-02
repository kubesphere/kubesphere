package http01

import (
	"fmt"
	"net/http"
	"strings"
)

// A domainMatcher tries to match a domain (the one we're requesting a certificate for)
// in the HTTP request coming from the ACME validation servers.
// This step is part of DNS rebind attack prevention,
// where the webserver matches incoming requests to a list of domain the server acts authoritative for.
//
// The most simple check involves finding the domain in the HTTP Host header;
// this is what hostMatcher does.
// Use it, when the http01.ProviderServer is directly reachable from the internet,
// or when it operates behind a transparent proxy.
//
// In many (reverse) proxy setups, Apache and NGINX traditionally move the Host header to a new header named X-Forwarded-Host.
// Use arbitraryMatcher("X-Forwarded-Host") in this case,
// or the appropriate header name for other proxy servers.
//
// RFC7239 has standardized the different forwarding headers into a single header named Forwarded.
// The header value has a different format, so you should use forwardedMatcher
// when the http01.ProviderServer operates behind a RFC7239 compatible proxy.
// https://tools.ietf.org/html/rfc7239
//
// Note: RFC7239 also reminds us, "that an HTTP list [...] may be split over multiple header fields" (section 7.1),
// meaning that
//   X-Header: a
//   X-Header: b
// is equal to
//   X-Header: a, b
//
// All matcher implementations (explicitly not excluding arbitraryMatcher!)
// have in common that they only match against the first value in such lists.
type domainMatcher interface {
	// matches checks whether the request is valid for the given domain.
	matches(request *http.Request, domain string) bool

	// name returns the header name used in the check.
	// This is primarily used to create meaningful error messages.
	name() string
}

// hostMatcher checks whether (*net/http).Request.Host starts with a domain name.
type hostMatcher struct{}

func (m *hostMatcher) name() string {
	return "Host"
}

func (m *hostMatcher) matches(r *http.Request, domain string) bool {
	return strings.HasPrefix(r.Host, domain)
}

// hostMatcher checks whether the specified (*net/http.Request).Header value starts with a domain name.
type arbitraryMatcher string

func (m arbitraryMatcher) name() string {
	return string(m)
}

func (m arbitraryMatcher) matches(r *http.Request, domain string) bool {
	return strings.HasPrefix(r.Header.Get(m.name()), domain)
}

// forwardedMatcher checks whether the Forwarded header contains a "host" element starting with a domain name.
// See https://tools.ietf.org/html/rfc7239 for details.
type forwardedMatcher struct{}

func (m *forwardedMatcher) name() string {
	return "Forwarded"
}

func (m *forwardedMatcher) matches(r *http.Request, domain string) bool {
	fwds, err := parseForwardedHeader(r.Header.Get(m.name()))
	if err != nil {
		return false
	}

	if len(fwds) == 0 {
		return false
	}

	host := fwds[0]["host"]
	return strings.HasPrefix(host, domain)
}

// parsing requires some form of state machine
func parseForwardedHeader(s string) (elements []map[string]string, err error) {
	cur := make(map[string]string)
	key := ""
	val := ""
	inquote := false

	pos := 0
	l := len(s)
	for i := 0; i < l; i++ {
		r := rune(s[i])

		if inquote {
			if r == '"' {
				cur[key] = s[pos:i]
				key = ""
				pos = i
				inquote = false
			}
			continue
		}

		switch {
		case r == '"': // start of quoted-string
			if key == "" {
				return nil, fmt.Errorf("unexpected quoted string as pos %d", i)
			}
			inquote = true
			pos = i + 1

		case r == ';': // end of forwarded-pair
			cur[key] = s[pos:i]
			key = ""
			i = skipWS(s, i)
			pos = i + 1

		case r == '=': // end of token
			key = strings.ToLower(strings.TrimFunc(s[pos:i], isWS))
			i = skipWS(s, i)
			pos = i + 1

		case r == ',': // end of forwarded-element
			if key != "" {
				if val == "" {
					val = s[pos:i]
				}
				cur[key] = val
			}
			elements = append(elements, cur)
			cur = make(map[string]string)
			key = ""
			val = ""

			i = skipWS(s, i)
			pos = i + 1
		case tchar(r) || isWS(r): // valid token character or whitespace
			continue
		default:
			return nil, fmt.Errorf("invalid token character at pos %d: %c", i, r)
		}
	}

	if inquote {
		return nil, fmt.Errorf("unterminated quoted-string at pos %d", len(s))
	}

	if key != "" {
		if pos < len(s) {
			val = s[pos:]
		}
		cur[key] = val
	}
	if len(cur) > 0 {
		elements = append(elements, cur)
	}
	return elements, nil
}

func tchar(r rune) bool {
	return strings.ContainsRune("!#$%&'*+-.^_`|~", r) ||
		'0' <= r && r <= '9' ||
		'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z'
}

func skipWS(s string, i int) int {
	for isWS(rune(s[i+1])) {
		i++
	}
	return i
}

func isWS(r rune) bool {
	return strings.ContainsRune(" \t\v\r\n", r)
}
