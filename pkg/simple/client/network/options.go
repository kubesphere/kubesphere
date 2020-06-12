package network

import "github.com/spf13/pflag"

type Options struct {
	EnableNetworkPolicy bool `json:"enableNetworkPolicy,omitempty" yaml:"enableNetworkPolicy"`
}

// NewNetworkOptions returns a `zero` instance
func NewNetworkOptions() *Options {
	return &Options{
		EnableNetworkPolicy: false,
	}
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) ApplyTo(options *Options) {
	options.EnableNetworkPolicy = s.EnableNetworkPolicy
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.EnableNetworkPolicy, "enable-network-policy", c.EnableNetworkPolicy,
		"This field instructs KubeSphere to enable network policy or not.")
}
