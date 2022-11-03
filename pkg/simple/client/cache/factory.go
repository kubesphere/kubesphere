package cache

type CacheFactory interface {
	// Type unique type of the cache
	Type() string
	// Create relevant caches by type
	Create(options DynamicOptions, stopCh <-chan struct{}) (Interface, error)
}
