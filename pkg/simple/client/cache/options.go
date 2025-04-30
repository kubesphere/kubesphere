/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cache

import (
	"fmt"

	"kubesphere.io/kubesphere/pkg/server/options"
)

type Options struct {
	Type    string                 `json:"type"`
	Options options.DynamicOptions `json:"options"`
}

// NewCacheOptions returns options points to nowhere,
// because redis is not required for some components
func NewCacheOptions() *Options {
	return &Options{
		Type:    TypeInMemoryCache,
		Options: map[string]interface{}{},
	}
}

// Validate check options
func (r *Options) Validate() []error {
	errors := make([]error, 0)

	if r.Type == "" {
		errors = append(errors, fmt.Errorf("invalid cache type"))
	}

	return errors
}
