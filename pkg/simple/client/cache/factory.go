package cache

import "kubesphere.io/kubesphere/pkg/server/options"

type CacheFactory interface {
	// Type unique type of the cache
	Type() string
	// Create relevant caches by type
	Create(options options.DynamicOptions, stopCh <-chan struct{}) (Interface, error)
}
