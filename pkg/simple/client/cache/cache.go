package cache

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"
)

var (
	cacheFactories = make(map[string]CacheFactory)
)

var NeverExpire = time.Duration(0)

type Interface interface {
	// Keys retrieves all keys match the given pattern
	Keys(pattern string) ([]string, error)

	// Get retrieves the value of the given key, return error if key doesn't exist
	Get(key string) (string, error)

	// Set sets the value and living duration of the given key, zero duration means never expire
	Set(key string, value string, duration time.Duration) error

	// Del deletes the given key, no error returned if the key doesn't exist
	Del(keys ...string) error

	// Exists checks the existence of a give key
	Exists(keys ...string) (bool, error)

	// Expire updates object's expiration time, return err if key doesn't exist
	Expire(key string, duration time.Duration) error
}

func RegisterCacheFactory(factory CacheFactory) {
	cacheFactories[factory.Type()] = factory
}

func New(option *Options, stopCh <-chan struct{}) (Interface, error) {
	if cacheFactories[option.Type] == nil {
		err := fmt.Errorf("cache with type %s is not supported", option.Type)
		klog.Error(err)
		return nil, err
	}

	if option.Type == TypeInMemoryCache {
		klog.Warning("In-memory cache will be used, this may cause data inconsistencies when running with multiple replicas.")
	}

	cache, err := cacheFactories[option.Type].Create(option.Options, stopCh)
	if err != nil {
		klog.Errorf("failed to create cache, error: %v", err)
		return nil, err
	}
	return cache, nil
}
