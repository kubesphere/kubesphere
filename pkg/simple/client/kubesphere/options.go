package kubesphere

import "github.com/spf13/pflag"

type Options struct {
	APIServer     string `json:"apiServer" yaml:"apiServer"`
	AccountServer string `json:"accountServer" yaml:"accountServer"`
}

// NewKubeSphereOptions create a default options
func NewKubeSphereOptions() *Options {
	return &Options{
		APIServer:     "http://ks-apiserver.kubesphere-system.svc",
		AccountServer: "http://ks-account.kubesphere-system.svc",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if s.AccountServer != "" {
		options.AccountServer = s.AccountServer
	}

	if s.APIServer != "" {
		options.APIServer = s.APIServer
	}
}

func (s *Options) Validate() []error {
	errs := []error{}

	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.APIServer, "kubesphere-apiserver-host", s.APIServer, ""+
		"KubeSphere apiserver host address.")

	fs.StringVar(&s.AccountServer, "kubesphere-account-host", s.AccountServer, ""+
		"KubeSphere account server host address.")
}
