package metrics

import "github.com/spf13/pflag"

type Options struct {
	Enable bool `json:"enable,omitempty" description:"enable metric"`
}

func NewMetricsOptions() *Options {
	return &Options{
		Enable: false,
	}
}

func (s *Options) ApplyTo(options *Options) {
	if options == nil {
		options = s
		return
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.Enable, "enable-metric", c.Enable, "If true, allow metric. [default=false]")
}
