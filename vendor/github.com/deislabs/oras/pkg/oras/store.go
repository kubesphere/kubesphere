package oras

import (
	"context"
	"errors"

	orascontent "github.com/deislabs/oras/pkg/content"

	"github.com/containerd/containerd/content"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ensure interface
var (
	_ content.Store = &hybridStore{}
)

type hybridStore struct {
	cache    *orascontent.Memorystore
	provider content.Provider
	ingester content.Ingester
}

func newHybridStoreFromProvider(provider content.Provider) *hybridStore {
	return &hybridStore{
		cache:    orascontent.NewMemoryStore(),
		provider: provider,
	}
}

func newHybridStoreFromIngester(ingester content.Ingester) *hybridStore {
	return &hybridStore{
		cache:    orascontent.NewMemoryStore(),
		ingester: ingester,
	}
}

func (s *hybridStore) Set(desc ocispec.Descriptor, content []byte) {
	s.cache.Set(desc, content)
}

// ReaderAt provides contents
func (s *hybridStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	readerAt, err := s.cache.ReaderAt(ctx, desc)
	if err == nil {
		return readerAt, nil
	}
	if s.provider != nil {
		return s.provider.ReaderAt(ctx, desc)
	}
	return nil, err
}

// Writer begins or resumes the active writer identified by desc
func (s *hybridStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}

	if isAllowedMediaType(wOpts.Desc.MediaType, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageIndex) || s.ingester == nil {
		return s.cache.Writer(ctx, opts...)
	}
	return s.ingester.Writer(ctx, opts...)
}

// TODO: implement (needed to create a content.Store)
// TODO: do not return empty content.Info
// Abort completely cancels the ingest operation targeted by ref.
func (s *hybridStore) Info(ctx context.Context, dgst digest.Digest) (content.Info, error) {
	return content.Info{}, nil
}

// TODO: implement (needed to create a content.Store)
// Update updates mutable information related to content.
// If one or more fieldpaths are provided, only those
// fields will be updated.
// Mutable fields:
//  labels.*
func (s *hybridStore) Update(ctx context.Context, info content.Info, fieldpaths ...string) (content.Info, error) {
	return content.Info{}, errors.New("not yet implemented: Update (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Walk will call fn for each item in the content store which
// match the provided filters. If no filters are given all
// items will be walked.
func (s *hybridStore) Walk(ctx context.Context, fn content.WalkFunc, filters ...string) error {
	return errors.New("not yet implemented: Walk (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Delete removes the content from the store.
func (s *hybridStore) Delete(ctx context.Context, dgst digest.Digest) error {
	return errors.New("not yet implemented: Delete (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
func (s *hybridStore) Status(ctx context.Context, ref string) (content.Status, error) {
	// Status returns the status of the provided ref.
	return content.Status{}, errors.New("not yet implemented: Status (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// ListStatuses returns the status of any active ingestions whose ref match the
// provided regular expression. If empty, all active ingestions will be
// returned.
func (s *hybridStore) ListStatuses(ctx context.Context, filters ...string) ([]content.Status, error) {
	return []content.Status{}, errors.New("not yet implemented: ListStatuses (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Abort completely cancels the ingest operation targeted by ref.
func (s *hybridStore) Abort(ctx context.Context, ref string) error {
	return errors.New("not yet implemented: Abort (content.Store interface)")
}
