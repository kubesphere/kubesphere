package cache

type fakeCacheFactory struct {
}

func (fc fakeCacheFactory) Type() string {
	return "FAKE"
}

func (fc *fakeCacheFactory) Create(options DynamicOptions, stopCh <-chan struct{}) (Interface, error) {
	return NewSimpleCache(nil, nil)
}

func init() {
	RegisterCacheFactory(&fakeCacheFactory{})
}
