package v4

import (
	"net/url"
	"strings"
)

const doubleSpace = "  "

// StripExcessSpaces will rewrite the passed in slice's string values to not
// contain multiple side-by-side spaces.
func StripExcessSpaces(str string) string {
	var j, k, l, m, spaces int

	// Trim leading and trailing spaces
	str = strings.Trim(str, " ")

	// Strip multiple spaces.
	j = strings.Index(str, doubleSpace)
	if j < 0 {
		return str
	}

	buf := []byte(str)
	for k, m, l = j, j, len(buf); k < l; k++ {
		if buf[k] == ' ' {
			if spaces == 0 {
				// First space.
				buf[m] = buf[k]
				m++
			}
			spaces++
		} else {
			// End of multiple spaces.
			spaces = 0
			buf[m] = buf[k]
			m++
		}
	}

	return string(buf[:m])
}

// GetURIPath returns the escaped URI component from the provided URL
func GetURIPath(u *url.URL) string {
	var uri string

	if len(u.Opaque) > 0 {
		uri = "/" + strings.Join(strings.Split(u.Opaque, "/")[3:], "/")
	} else {
		uri = u.EscapedPath()
	}

	if len(uri) == 0 {
		uri = "/"
	}

	return uri
}
