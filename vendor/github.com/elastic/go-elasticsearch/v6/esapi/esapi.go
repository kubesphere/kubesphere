package esapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v6/internal/version"
)

// Version returns the package version as a string.
//
const Version = version.Client

// Transport defines the interface for an API client.
//
type Transport interface {
	Perform(*http.Request) (*http.Response, error)
}

// BoolPtr returns a pointer to v.
//
// It is used as a convenience function for converting a bool value
// into a pointer when passing the value to a function or struct field
// which expects a pointer.
//
func BoolPtr(v bool) *bool { return &v }

// IntPtr returns a pointer to v.
//
// It is used as a convenience function for converting an int value
// into a pointer when passing the value to a function or struct field
// which expects a pointer.
//
func IntPtr(v int) *int { return &v }

// formatDuration converts duration to a string in the format
// accepted by Elasticsearch.
//
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return strconv.FormatInt(int64(d), 10) + "nanos"
	}
	return strconv.FormatInt(int64(d)/int64(time.Millisecond), 10) + "ms"
}
