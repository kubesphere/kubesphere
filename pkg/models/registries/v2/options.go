/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"context"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	// DefaultRegistry is the registry name that will be used if no registry
	// provided and the default is not overridden.
	DefaultRegistry = "index.docker.io"

	// DefaultTag is the tag name that will be used if no tag provided and the
	// default is not overridden.
	DefaultTag = "latest"
)

type options struct {
	name     []name.Option
	remote   []remote.Option
	platform *v1.Platform
}

func makeOptions(opts ...Option) options {
	opt := options{
		remote: []remote.Option{
			remote.WithAuth(authn.Anonymous),
		},
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// Option is a functional option
type Option func(*options)

// WithTransport is a functional option for overriding the default transport
// for remote operations.
func WithTransport(t http.RoundTripper) Option {
	return func(o *options) {
		o.remote = append(o.remote, remote.WithTransport(t))
	}
}

// Insecure is an Option that allows image references to be fetched without TLS.
func Insecure(o *options) {
	o.name = append(o.name, name.Insecure)
}

// WithAuth is a functional option for overriding the default authenticator
// for remote operations.
func WithAuth(auth authn.Authenticator) Option {
	return func(o *options) {
		// Replace the default keychain at position 0.
		o.remote[0] = remote.WithAuth(auth)
	}
}

// WithContext is a functional option for setting the context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.remote = append(o.remote, remote.WithContext(ctx))
	}
}

// WithPlatform is an Option to specify the platform.
func WithPlatform(platform *v1.Platform) Option {
	return func(o *options) {
		if platform != nil {
			o.remote = append(o.remote, remote.WithPlatform(*platform))
		}
		o.platform = platform
	}
}
