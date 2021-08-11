package oras

import (
	"context"

	orascontent "github.com/deislabs/oras/pkg/content"

	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/semaphore"
)

type pullOpts struct {
	allowedMediaTypes      []string
	dispatch               func(context.Context, images.Handler, *semaphore.Weighted, ...ocispec.Descriptor) error
	baseHandlers           []images.Handler
	callbackHandlers       []images.Handler
	contentProvideIngester orascontent.ProvideIngester
	filterName             func(ocispec.Descriptor) bool
}

// PullOpt allows callers to set options on the oras pull
type PullOpt func(o *pullOpts) error

func pullOptsDefaults() *pullOpts {
	return &pullOpts{
		dispatch:   images.Dispatch,
		filterName: filterName,
	}
}

// WithAllowedMediaType sets the allowed media types
func WithAllowedMediaType(allowedMediaTypes ...string) PullOpt {
	return func(o *pullOpts) error {
		o.allowedMediaTypes = append(o.allowedMediaTypes, allowedMediaTypes...)
		return nil
	}
}

// WithAllowedMediaTypes sets the allowed media types
func WithAllowedMediaTypes(allowedMediaTypes []string) PullOpt {
	return func(o *pullOpts) error {
		o.allowedMediaTypes = append(o.allowedMediaTypes, allowedMediaTypes...)
		return nil
	}
}

// WithPullByBFS opt to pull in sequence with breath-first search
func WithPullByBFS(o *pullOpts) error {
	o.dispatch = dispatchBFS
	return nil
}

// WithPullBaseHandler provides base handlers, which will be called before
// any pull specific handlers.
func WithPullBaseHandler(handlers ...images.Handler) PullOpt {
	return func(o *pullOpts) error {
		o.baseHandlers = append(o.baseHandlers, handlers...)
		return nil
	}
}

// WithPullCallbackHandler provides callback handlers, which will be called after
// any pull specific handlers.
func WithPullCallbackHandler(handlers ...images.Handler) PullOpt {
	return func(o *pullOpts) error {
		o.callbackHandlers = append(o.callbackHandlers, handlers...)
		return nil
	}
}

// WithContentProvideIngester opt to the provided Provider and Ingester
// for file system I/O, including caches.
func WithContentProvideIngester(store orascontent.ProvideIngester) PullOpt {
	return func(o *pullOpts) error {
		o.contentProvideIngester = store
		return nil
	}
}

// WithPullEmptyNameAllowed allows pulling blobs with empty name.
func WithPullEmptyNameAllowed() PullOpt {
	return func(o *pullOpts) error {
		o.filterName = func(ocispec.Descriptor) bool {
			return true
		}
		return nil
	}
}
