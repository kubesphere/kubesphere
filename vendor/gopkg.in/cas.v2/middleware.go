package cas

import (
	"net/http"

	"github.com/golang/glog"
)

// Handler returns a standard http.HandlerFunc, which will check the authenticated status (redirect user go login if needed)
// If the user pass the authenticated check, it will call the h's ServeHTTP method
func (c *Client) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if glog.V(2) {
			glog.Infof("cas: handling %v request for %v", r.Method, r.URL)
		}

		setClient(r, c)

		if !IsAuthenticated(r) {
			RedirectToLogin(w, r)
			return
		}

		if r.URL.Path == "/logout" {
			RedirectToLogout(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
