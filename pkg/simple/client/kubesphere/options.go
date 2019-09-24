package kubesphere

import "github.com/spf13/pflag"

type KubeSphereOptions struct {
	APIServer     string `json:"apiServer" yaml:"apiServer"`
	AccountServer string `json:"accountServer" yaml:"accountServer"`
}

// NewKubeSphereOptions create a default options
func NewKubeSphereOptions() *KubeSphereOptions {
	return &KubeSphereOptions{
		APIServer:     "http://ks-apiserver.kubesphere-system.svc",
		AccountServer: "http://ks-account.kubesphere-system.svc",
	}
}

func (s *KubeSphereOptions) ApplyTo(options *KubeSphereOptions) {
	if s.AccountServer != "" {
		options.AccountServer = s.AccountServer
	}

	if s.APIServer != "" {
		options.APIServer = s.APIServer
	}
}

func (s *KubeSphereOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (s *KubeSphereOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.APIServer, "kubesphere-apiserver-host", s.APIServer, ""+
		"KubeSphere apiserver host address.")

	fs.StringVar(&s.AccountServer, "kubesphere-account-host", s.AccountServer, ""+
		"KubeSphere account server host address.")
}
