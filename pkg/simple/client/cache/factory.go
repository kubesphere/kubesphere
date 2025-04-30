/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cache

import "kubesphere.io/kubesphere/pkg/server/options"

type CacheFactory interface {
	// Type unique type of the cache
	Type() string
	// Create relevant caches by type
	Create(options options.DynamicOptions, stopCh <-chan struct{}) (Interface, error)
}
