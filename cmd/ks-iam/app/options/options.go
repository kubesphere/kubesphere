package options

import (
	"github.com/spf13/pflag"
	genericoptions "kubesphere.io/kubesphere/pkg/options"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
	}

	return s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	s.GenericServerRunOptions.AddFlags(fs)
}
