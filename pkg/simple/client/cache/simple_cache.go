package cache

import "time"

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

func (s *SimpleCache) Del(keys ...string) error {
	panic("implement me")
}

func (s *SimpleCache) Get(key string) (string, error) {
	return "", nil
}

func (s *SimpleCache) Exists(keys ...string) (bool, error) {
	panic("implement me")
}

func (s *SimpleCache) Expire(key string, duration time.Duration) error {
	panic("implement me")
}
