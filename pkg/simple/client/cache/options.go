package cache

import (
	"fmt"
	"github.com/spf13/pflag"
)

type Options struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// NewRedisOptions returns options points to nowhere,
// because redis is not required for some components
func NewRedisOptions() *Options {
	return &Options{
		Host:     "",
		Port:     0,
		Password: "",
		DB:       0,
	}
}

// Validate check options
func (r *Options) Validate() []error {
	errors := make([]error, 0)

	if r.Port == 0 {
		errors = append(errors, fmt.Errorf("invalid service port number"))
	}

	return errors
}

// AddFlags add option flags to command line flags,
// if redis-host left empty, the following options will be ignored.
func (r *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&r.Host, "redis-host", s.Host, "Redis connection URL. If left blank, means redis is unnecessary, "+
		"redis will be disabled.")

	fs.IntVar(&r.Port, "redis-port", s.Port, "")
	fs.StringVar(&r.Password, "redis-password", s.Password, "")
	fs.IntVar(&r.DB, "redis-db", s.DB, "")
}
