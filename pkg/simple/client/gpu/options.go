package gpu

import "github.com/spf13/pflag"

type GPUKind struct {
	ResourceName string `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	ResourceType string `json:"resourceType,omitempty" yaml:"resourceType,omitempty"`
	Default      bool   `json:"default,omitempty" yaml:"default,omitempty"`
}

type Options struct {
	Kinds []GPUKind `json:"kinds,omitempty" yaml:"kinds,omitempty"`
}

func NewGPUOptions() *Options {
	return &Options{
		Kinds: []GPUKind{},
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {
	if s != nil && len(s.Kinds) > 0 {
		options.Kinds = s.Kinds
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {

}
