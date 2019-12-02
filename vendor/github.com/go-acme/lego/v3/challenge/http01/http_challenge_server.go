package http01

import (
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/go-acme/lego/v3/log"
)

// ProviderServer implements ChallengeProvider for `http-01` challenge
// It may be instantiated without using the NewProviderServer function if
// you want only to use the default values.
type ProviderServer struct {
	iface    string
	port     string
	matcher  domainMatcher
	done     chan bool
	listener net.Listener
}

// NewProviderServer creates a new ProviderServer on the selected interface and port.
// Setting iface and / or port to an empty string will make the server fall back to
// the "any" interface and port 80 respectively.
func NewProviderServer(iface, port string) *ProviderServer {
	if port == "" {
		port = "80"
	}

	return &ProviderServer{iface: iface, port: port, matcher: &hostMatcher{}}
}

// Present starts a web server and makes the token available at `ChallengePath(token)` for web requests.
func (s *ProviderServer) Present(domain, token, keyAuth string) error {
	var err error
	s.listener, err = net.Listen("tcp", s.GetAddress())
	if err != nil {
		return fmt.Errorf("could not start HTTP server for challenge -> %v", err)
	}

	s.done = make(chan bool)
	go s.serve(domain, token, keyAuth)
	return nil
}

func (s *ProviderServer) GetAddress() string {
	return net.JoinHostPort(s.iface, s.port)
}

// CleanUp closes the HTTP server and removes the token from `ChallengePath(token)`
func (s *ProviderServer) CleanUp(domain, token, keyAuth string) error {
	if s.listener == nil {
		return nil
	}
	s.listener.Close()
	<-s.done
	return nil
}

// SetProxyHeader changes the validation of incoming requests.
// By default, s matches the "Host" header value to the domain name.
//
// When the server runs behind a proxy server, this is not the correct place to look at;
// Apache and NGINX have traditionally moved the original Host header into a new header named "X-Forwarded-Host".
// Other webservers might use different names;
// and RFC7239 has standadized a new header named "Forwarded" (with slightly different semantics).
//
// The exact behavior depends on the value of headerName:
// - "" (the empty string) and "Host" will restore the default and only check the Host header
// - "Forwarded" will look for a Forwarded header, and inspect it according to https://tools.ietf.org/html/rfc7239
// - any other value will check the header value with the same name
func (s *ProviderServer) SetProxyHeader(headerName string) {
	switch h := textproto.CanonicalMIMEHeaderKey(headerName); h {
	case "", "Host":
		s.matcher = &hostMatcher{}
	case "Forwarded":
		s.matcher = &forwardedMatcher{}
	default:
		s.matcher = arbitraryMatcher(h)
	}
}

func (s *ProviderServer) serve(domain, token, keyAuth string) {
	path := ChallengePath(token)

	// The incoming request must will be validated to prevent DNS rebind attacks.
	// We only respond with the keyAuth, when we're receiving a GET requests with
	// the "Host" header matching the domain (the latter is configurable though SetProxyHeader).
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && s.matcher.matches(r, domain) {
			w.Header().Add("Content-Type", "text/plain")
			_, err := w.Write([]byte(keyAuth))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Infof("[%s] Served key authentication", domain)
		} else {
			log.Warnf("Received request for domain %s with method %s but the domain did not match any challenge. Please ensure your are passing the %s header properly.", r.Host, r.Method, s.matcher.name())
			_, err := w.Write([]byte("TEST"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	httpServer := &http.Server{Handler: mux}

	// Once httpServer is shut down
	// we don't want any lingering connections, so disable KeepAlives.
	httpServer.SetKeepAlivesEnabled(false)

	err := httpServer.Serve(s.listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Println(err)
	}
	s.done <- true
}
