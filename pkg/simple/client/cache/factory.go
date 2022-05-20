package cache

type CacheFactory interface {
	// Type unique type of the provider
	Type() string
	Create(options DynamicOptions, stopCh <-chan struct{}) (Interface, error)
}
