// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package cache defines the inter-query cache interface that can cache data across queries
package cache

import (
	"container/list"
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/util"
)

const (
	defaultMaxSizeBytes                      = int64(0)   // unlimited
	defaultForcedEvictionThresholdPercentage = int64(100) // trigger at max_size_bytes
	defaultStaleEntryEvictionPeriodSeconds   = int64(0)   // never
)

// Config represents the configuration of the inter-query cache.
type Config struct {
	InterQueryBuiltinCache InterQueryBuiltinCacheConfig `json:"inter_query_builtin_cache"`
}

// InterQueryBuiltinCacheConfig represents the configuration of the inter-query cache that built-in functions can utilize.
// MaxSizeBytes - max capacity of cache in bytes
// ForcedEvictionThresholdPercentage - capacity usage in percentage after which forced FIFO eviction starts
// StaleEntryEvictionPeriodSeconds - time period between end of previous and start of new stale entry eviction routine
type InterQueryBuiltinCacheConfig struct {
	MaxSizeBytes                      *int64 `json:"max_size_bytes,omitempty"`
	ForcedEvictionThresholdPercentage *int64 `json:"forced_eviction_threshold_percentage,omitempty"`
	StaleEntryEvictionPeriodSeconds   *int64 `json:"stale_entry_eviction_period_seconds,omitempty"`
}

// ParseCachingConfig returns the config for the inter-query cache.
func ParseCachingConfig(raw []byte) (*Config, error) {
	if raw == nil {
		maxSize := new(int64)
		*maxSize = defaultMaxSizeBytes
		threshold := new(int64)
		*threshold = defaultForcedEvictionThresholdPercentage
		period := new(int64)
		*period = defaultStaleEntryEvictionPeriodSeconds
		return &Config{InterQueryBuiltinCache: InterQueryBuiltinCacheConfig{MaxSizeBytes: maxSize, ForcedEvictionThresholdPercentage: threshold, StaleEntryEvictionPeriodSeconds: period}}, nil
	}

	var config Config

	if err := util.Unmarshal(raw, &config); err == nil {
		if err = config.validateAndInjectDefaults(); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validateAndInjectDefaults() error {
	if c.InterQueryBuiltinCache.MaxSizeBytes == nil {
		maxSize := new(int64)
		*maxSize = defaultMaxSizeBytes
		c.InterQueryBuiltinCache.MaxSizeBytes = maxSize
	}
	if c.InterQueryBuiltinCache.ForcedEvictionThresholdPercentage == nil {
		threshold := new(int64)
		*threshold = defaultForcedEvictionThresholdPercentage
		c.InterQueryBuiltinCache.ForcedEvictionThresholdPercentage = threshold
	} else {
		threshold := *c.InterQueryBuiltinCache.ForcedEvictionThresholdPercentage
		if threshold < 0 || threshold > 100 {
			return fmt.Errorf("invalid forced_eviction_threshold_percentage %v", threshold)
		}
	}
	if c.InterQueryBuiltinCache.StaleEntryEvictionPeriodSeconds == nil {
		period := new(int64)
		*period = defaultStaleEntryEvictionPeriodSeconds
		c.InterQueryBuiltinCache.StaleEntryEvictionPeriodSeconds = period
	} else {
		period := *c.InterQueryBuiltinCache.StaleEntryEvictionPeriodSeconds
		if period < 0 {
			return fmt.Errorf("invalid stale_entry_eviction_period_seconds %v", period)
		}
	}
	return nil
}

// InterQueryCacheValue defines the interface for the data that the inter-query cache holds.
type InterQueryCacheValue interface {
	SizeInBytes() int64
	Clone() (InterQueryCacheValue, error)
}

// InterQueryCache defines the interface for the inter-query cache.
type InterQueryCache interface {
	Get(key ast.Value) (value InterQueryCacheValue, found bool)
	Insert(key ast.Value, value InterQueryCacheValue) int
	InsertWithExpiry(key ast.Value, value InterQueryCacheValue, expiresAt time.Time) int
	Delete(key ast.Value)
	UpdateConfig(config *Config)
	Clone(value InterQueryCacheValue) (InterQueryCacheValue, error)
}

// NewInterQueryCache returns a new inter-query cache.
// The cache uses a FIFO eviction policy when it reaches the forced eviction threshold.
// Parameters:
//
//	config - to configure the InterQueryCache
func NewInterQueryCache(config *Config) InterQueryCache {
	return newCache(config)
}

// NewInterQueryCacheWithContext returns a new inter-query cache with context.
// The cache uses a combination of FIFO eviction policy when it reaches the forced eviction threshold
// and a periodic cleanup routine to remove stale entries that exceed their expiration time, if specified.
// If configured with a zero stale_entry_eviction_period_seconds value, the stale entry cleanup routine is disabled.
//
// Parameters:
//
//	ctx - used to control lifecycle of the stale entry cleanup routine
//	config - to configure the InterQueryCache
func NewInterQueryCacheWithContext(ctx context.Context, config *Config) InterQueryCache {
	iqCache := newCache(config)
	if iqCache.staleEntryEvictionTimePeriodSeconds() > 0 {
		cleanupTicker := time.NewTicker(time.Duration(iqCache.staleEntryEvictionTimePeriodSeconds()) * time.Second)
		go func() {
			for {
				select {
				case <-cleanupTicker.C:
					cleanupTicker.Stop()
					iqCache.cleanStaleValues()
					cleanupTicker = time.NewTicker(time.Duration(iqCache.staleEntryEvictionTimePeriodSeconds()) * time.Second)
				case <-ctx.Done():
					cleanupTicker.Stop()
					return
				}
			}
		}()
	}

	return iqCache
}

type cacheItem struct {
	value      InterQueryCacheValue
	expiresAt  time.Time
	keyElement *list.Element
}

type cache struct {
	items  map[string]cacheItem
	usage  int64
	config *Config
	l      *list.List
	mtx    sync.Mutex
}

func newCache(config *Config) *cache {
	return &cache{
		items:  map[string]cacheItem{},
		usage:  0,
		config: config,
		l:      list.New(),
	}
}

// InsertWithExpiry inserts a key k into the cache with value v with an expiration time expiresAt.
// A zero time value for expiresAt indicates no expiry
func (c *cache) InsertWithExpiry(k ast.Value, v InterQueryCacheValue, expiresAt time.Time) (dropped int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.unsafeInsert(k, v, expiresAt)
}

// Insert inserts a key k into the cache with value v with no expiration time.
func (c *cache) Insert(k ast.Value, v InterQueryCacheValue) (dropped int) {
	return c.InsertWithExpiry(k, v, time.Time{})
}

// Get returns the value in the cache for k.
func (c *cache) Get(k ast.Value) (InterQueryCacheValue, bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	cacheItem, ok := c.unsafeGet(k)

	if ok {
		return cacheItem.value, true
	}
	return nil, false
}

// Delete deletes the value in the cache for k.
func (c *cache) Delete(k ast.Value) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.unsafeDelete(k)
}

func (c *cache) UpdateConfig(config *Config) {
	if config == nil {
		return
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.config = config
}

func (c *cache) Clone(value InterQueryCacheValue) (InterQueryCacheValue, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.unsafeClone(value)
}

func (c *cache) unsafeInsert(k ast.Value, v InterQueryCacheValue, expiresAt time.Time) (dropped int) {
	size := v.SizeInBytes()
	limit := int64(math.Ceil(float64(c.forcedEvictionThresholdPercentage())/100.0) * (float64(c.maxSizeBytes())))
	if limit > 0 {
		if size > limit {
			dropped++
			return dropped
		}

		for key := c.l.Front(); key != nil && (c.usage+size > limit); key = c.l.Front() {
			dropKey := key.Value.(ast.Value)
			c.unsafeDelete(dropKey)
			dropped++
		}
	}

	// By deleting the old value, if it exists, we ensure the usage variable stays correct
	c.unsafeDelete(k)

	c.items[k.String()] = cacheItem{
		value:      v,
		expiresAt:  expiresAt,
		keyElement: c.l.PushBack(k),
	}
	c.usage += size
	return dropped
}

func (c *cache) unsafeGet(k ast.Value) (cacheItem, bool) {
	value, ok := c.items[k.String()]
	return value, ok
}

func (c *cache) unsafeDelete(k ast.Value) {
	cacheItem, ok := c.unsafeGet(k)
	if !ok {
		return
	}

	c.usage -= cacheItem.value.SizeInBytes()
	delete(c.items, k.String())
	c.l.Remove(cacheItem.keyElement)
}

func (c *cache) unsafeClone(value InterQueryCacheValue) (InterQueryCacheValue, error) {
	return value.Clone()
}

func (c *cache) maxSizeBytes() int64 {
	if c.config == nil {
		return defaultMaxSizeBytes
	}
	return *c.config.InterQueryBuiltinCache.MaxSizeBytes
}

func (c *cache) forcedEvictionThresholdPercentage() int64 {
	if c.config == nil {
		return defaultForcedEvictionThresholdPercentage
	}
	return *c.config.InterQueryBuiltinCache.ForcedEvictionThresholdPercentage
}

func (c *cache) staleEntryEvictionTimePeriodSeconds() int64 {
	if c.config == nil {
		return defaultStaleEntryEvictionPeriodSeconds
	}
	return *c.config.InterQueryBuiltinCache.StaleEntryEvictionPeriodSeconds
}

func (c *cache) cleanStaleValues() (dropped int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for key := c.l.Front(); key != nil; {
		nextKey := key.Next()
		// if expiresAt is zero, the item doesn't have an expiry
		if ea := c.items[(key.Value.(ast.Value)).String()].expiresAt; !ea.IsZero() && ea.Before(time.Now()) {
			c.unsafeDelete(key.Value.(ast.Value))
			dropped++
		}
		key = nextKey
	}
	return dropped
}
