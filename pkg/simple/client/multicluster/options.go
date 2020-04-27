package multicluster

import "github.com/spf13/pflag"

type Options struct {
	// Enable
	Enable           bool `json:"enable"`
	EnableFederation bool `json:"enableFederation,omitempty"`
}

// NewOptions() returns a default nil options
func NewOptions() *Options {
	return &Options{
		Enable:           false,
		EnableFederation: false,
	}
}

func (o *Options) Validate() []error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.BoolVar(&o.Enable, "multiple-clusters", s.Enable, ""+
		"This field instructs KubeSphere to enter multiple-cluster mode or not.")
}
