package names

import (
	"os"
	"strings"
)

// Hostname returns this hosts hostname, converting to lowercase so that
// it is valid for use in the Calico API.
func Hostname() (string, error) {
	if h, err := os.Hostname(); err != nil {
		return "", err
	} else {
		return strings.ToLower(strings.TrimSpace(h)), nil
	}
}
