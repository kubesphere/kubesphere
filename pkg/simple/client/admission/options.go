package admission

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Enable                   bool `json:"enable" yaml:"enable"`
	EnableGatekeeperProvider bool `json:"enableGatekeeperProvider" yaml:"enableGatekeeperProvider"`
}

func NewAdmissionOptions() *Options {
	return &Options{}
}

func (o *Options) ApplyTo(options *Options) {
	reflectutils.Override(options, o)
}

func (o *Options) Validate() []error {
	var errs []error

	return errs
}

func (o *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&o.Enable, "admission-enable", c.Enable,
		"Enable admission component or not.")
}
