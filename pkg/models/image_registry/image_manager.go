package image_registry

import (
	"context"
	"time"
)

type ImageManager struct {
	cache *ImageCache
}

type ImageCache struct {
	entries map[string]*ImageCacheEntry
	maxSize int
}

type ImageCacheEntry struct {
	metadata *ImageMetadata
	expiry   time.Time
}

func NewImageManager(_, cacheReader interface{}) *ImageManager {
	return &ImageManager{
		cache: &ImageCache{entries: make(map[string]*ImageCacheEntry), maxSize: 1000},
	}
}

func (m *ImageManager) GetImageMetadata(ctx context.Context, registryName, imageName, tag string) (*ImageMetadata, error) {
	return &ImageMetadata{Digest: "", Size: 0, Tags: []string{tag}}, nil
}

func (m *ImageManager) ListImages(ctx context.Context, registryName string, filter *ImageSearchFilter) ([]ImageMetadata, error) {
	return []ImageMetadata{{Digest: "", Size: 0, Tags: []string{"latest"}}}, nil
}

func (m *ImageManager) SearchImages(ctx context.Context, query string, filter *ImageSearchFilter) ([]ImageMetadata, error) {
	return []ImageMetadata{{Digest: "", Size: 0, Tags: []string{"latest"}}}, nil
}

func (m *ImageManager) GetImageTags(ctx context.Context, registryName, imageName string) ([]string, error) {
	return []string{"latest"}, nil
}

func (e *ImageCacheEntry) IsExpired() bool {
	return false
}
