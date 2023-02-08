package filters

import "net/http"

type Middleware interface {
	Handle(w http.ResponseWriter, req *http.Request) bool
}

func WithMiddleware(next http.Handler, middlewares ...Middleware) http.Handler {
	if middlewares == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, middleware := range middlewares {
			if middleware.Handle(w, req) {
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}
