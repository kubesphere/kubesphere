package options

import (
	"github.com/spf13/pflag"
	genericoptions "kubesphere.io/kubesphere/pkg/options"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	AdminEmail              string
	AdminPassword           string
	TokenExpireTime         string
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
	}
	return s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.AdminEmail, "admin-email", "admin@kubesphere.io", "default administrator's email")
	fs.StringVar(&s.AdminPassword, "admin-password", "passw0rd", "default administrator's password")
	fs.StringVar(&s.TokenExpireTime, "token-expire-time", "24h", "token expire time")
	s.GenericServerRunOptions.AddFlags(fs)
}
