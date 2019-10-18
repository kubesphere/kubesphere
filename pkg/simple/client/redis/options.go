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
