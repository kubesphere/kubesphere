/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redis

import (
	"github.com/go-redis/redis"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type RedisOptions struct {
	RedisURL string
}

// NewRedisOptions returns options points to nowhere,
// because redis is not required for some components
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		RedisURL: "",
	}
}

// Validate check options
func (r *RedisOptions) Validate() []error {
	errors := make([]error, 0)

	_, err := redis.ParseURL(r.RedisURL)

	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

// ApplyTo apply to another options if it's a enabled option(non empty host)
func (r *RedisOptions) ApplyTo(options *RedisOptions) {
	if r.RedisURL != "" {
		reflectutils.Override(options, r)
	}
}

// AddFlags add option flags to command line flags,
// if redis-host left empty, the following options will be ignored.
func (r *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.RedisURL, "redis-url", "", "Redis connection URL. If left blank, means redis is unnecessary, "+
		"redis will be disabled. e.g. redis://:password@host:port/db")
}
