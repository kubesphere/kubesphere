package cache

import "time"

type Interface interface {
	// Keys retrieves all keys match the given pattern
	Keys(pattern string) ([]string, error)

	// Get retrieves the value of the given key, return error if key doesn't exist
	Get(key string) (string, error)

	// Set sets the value and living duration of the given key, zero duration means never expire
	Set(key string, value string, duration time.Duration) error

	// Del deletes the given key, no error returned if the key doesn't exists
	Del(key string) error

	// Exists checks the existence of a give key
	Exists(key string) (bool, error)

	// Expires updates object's expiration time, return err if key doesn't exist
	Expire(key string, duration time.Duration) error
}

type simpleObject struct {
	value  string
	expire time.Time
}

type SimpleCache struct {
	store map[string]simpleObject
}

func NewSimpleCache() Interface {
	return &SimpleCache{store: make(map[string]simpleObject)}
}

func (s *SimpleCache) Keys(pattern string) ([]string, error) {
	panic("implement me")
}

func (s *SimpleCache) Set(key string, value string, duration time.Duration) error {
	panic("implement me")
}

func (s *SimpleCache) Del(key string) error {
	panic("implement me")
}

func (s *SimpleCache) Get(key string) (string, error) {
	return "", nil
}

func (s *SimpleCache) Exists(key string) (bool, error) {
	panic("implement me")
}

func (s *SimpleCache) Expire(key string, duration time.Duration) error {
	panic("implement me")
}
