package openpitrix

import (
	"fmt"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type OpenPitrixOptions struct {
	APIServer string `json:"apiServer,omitempty" yaml:"apiServer,omitempty"`
	Token     string `json:"token,omitempty" yaml:"token,omitempty"`
}

func NewOpenPitrixOptions() *OpenPitrixOptions {
	return &OpenPitrixOptions{
		APIServer: "",
		Token:     "",
	}
}

func (s *OpenPitrixOptions) ApplyTo(options *OpenPitrixOptions) {
	reflectutils.Override(s, options)
}

func (s *OpenPitrixOptions) Validate() []error {
	errs := []error{}

	if s.APIServer != "" {
		if s.Token == "" {
			errs = append(errs, fmt.Errorf("OpenPitrix access token cannot be empty"))
		}
	}

	return errs
}

func (s *OpenPitrixOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.APIServer, "openpitrix-apiserver", s.APIServer, ""+
		"OpenPitrix api gateway endpoint, if left blank, following options will be ignored.")

	fs.StringVar(&s.Token, "openpitrix-token", s.Token, ""+
		"OpenPitrix api access token.")
}
