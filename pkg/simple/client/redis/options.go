package redis

import (
	"fmt"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/net"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type RedisOptions struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisOptions returns options points to nowhere,
// because redis is not required for some components
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Host:     "",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

// Validate check options
func (r *RedisOptions) Validate() []error {
	errors := make([]error, 0)

	if r.Host != "" {
		if !net.IsValidPort(r.Port) {
			errors = append(errors, fmt.Errorf("--redis-port is out of range"))
		}
	}

	if r.DB < 0 {
		errors = append(errors, fmt.Errorf("--redis-db is less than 0"))
	}

	return errors
}

// ApplyTo apply to another options if it's a enabled option(non empty host)
func (r *RedisOptions) ApplyTo(options *RedisOptions) {
	if r.Host != "" {
		reflectutils.Override(options, r)
	}
}

// AddFlags add option flags to command line flags,
// if redis-host left empty, the following options will be ignored.
func (r *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.Host, "redis-host", r.Host, ""+
		"Redis service host address. If left blank, means redis is unnecessary, "+
		"redis will be disabled")

	fs.IntVar(&r.Port, "redis-port", r.Port, ""+
		"Redis service port number.")

	fs.StringVar(&r.Password, "redis-password", r.Password, ""+
		"Redis service password if necessary, default to empty")

	fs.IntVar(&r.DB, "redis-db", r.DB, ""+
		"Redis service database index, default to 0.")
}
