// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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
	DefaultRegistry      = "index.docker.io"
	defaultRegistryAlias = "docker.io"

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
//
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
