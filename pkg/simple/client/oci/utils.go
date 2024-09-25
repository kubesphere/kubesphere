package oci

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// defaultMaxMetadataBytes specifies the default limit on how many response
// bytes are allowed in the server's response to the metadata APIs.
// See also: Repository.MaxMetadataBytes
var defaultMaxMetadataBytes int64 = 4 * 1024 * 1024 // 4 MiB

// errNoLink is returned by parseLink() when no Link header is present.
var errNoLink = errors.New("no Link header in response")

// parseLink returns the URL of the response's "Link" header, if present.
func parseLink(resp *http.Response) (string, error) {
	link := resp.Header.Get("Link")
	if link == "" {
		return "", errNoLink
	}
	if link[0] != '<' {
		return "", fmt.Errorf("invalid next link %q: missing '<'", link)
	}
	if i := strings.IndexByte(link, '>'); i == -1 {
		return "", fmt.Errorf("invalid next link %q: missing '>'", link)
	} else {
		link = link[1:i]
	}

	linkURL, err := resp.Request.URL.Parse(link)
	if err != nil {
		return "", err
	}
	return linkURL.String(), nil
}

// limitReader returns a Reader that reads from r but stops with EOF after n
// bytes. If n is zero, defaultMaxMetadataBytes is used.
func limitReader(r io.Reader, n int64) io.Reader {
	if n == 0 {
		n = defaultMaxMetadataBytes
	}
	return io.LimitReader(r, n)
}
