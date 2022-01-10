package terminal

import "github.com/spf13/pflag"

type Options struct {
	Image   string `json:"image,omitempty" yaml:"image"`
	Timeout int    `json:"timeout,omitempty" yaml:"timeout"`
}

func NewTerminalOptions() *Options {
	return &Options{
		Image:   "alpine:3.15",
		Timeout: 600,
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {

}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {

}
