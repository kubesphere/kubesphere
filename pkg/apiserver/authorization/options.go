package authorization

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type Options struct {
	Mode string `json:"mode" yaml:"mode"`
}

func NewOptions() *Options {
	return &Options{Mode: RBAC}
}

var (
	AlwaysDeny  = "AlwaysDeny"
	AlwaysAllow = "AlwaysAllow"
	RBAC        = "RBAC"
)

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&o.Mode, "authorization", s.Mode, "Authorization setting, allowed values: AlwaysDeny, AlwaysAllow, RBAC.")
}

func (o *Options) Validate() []error {
	errs := make([]error, 0)
	if !sliceutil.HasString([]string{AlwaysAllow, AlwaysDeny, RBAC}, o.Mode) {
		err := fmt.Errorf("authorization mode %s not support", o.Mode)
		klog.Error(err)
		errs = append(errs, err)
	}
	return errs
}
