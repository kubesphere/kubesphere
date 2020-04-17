package network

import "github.com/spf13/pflag"

type Options struct {

	// weave scope service host
	WeaveScopeHost string `json:"weaveScopeHost,omitempty" yaml:"weaveScopeHost"`
}

// NewNetworkOptions returns a `zero` instance
func NewNetworkOptions() *Options {
	return &Options{
		WeaveScopeHost: "weave-scope-app.weave.svc",
	}
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) ApplyTo(options *Options) {
	if s.WeaveScopeHost != "" {
		options.WeaveScopeHost = s.WeaveScopeHost
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.WeaveScopeHost, "weave-scope-host", c.WeaveScopeHost, ""+
		"weave scope service host")
}
