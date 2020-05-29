package sonarqube

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Host  string `json:",omitempty" yaml:"host" description:"SonarQube service host address"`
	Token string `json:",omitempty" yaml:"token" description:"SonarQube service token"`
}

func NewSonarQubeOptions() *Options {
	return &Options{
		Host:  "",
		Token: "",
	}
}

func (s *Options) Validate() []error {
	var errors []error

	return errors
}

func (s *Options) ApplyTo(options *Options) {
	if s.Host != "" {
		options.Host = s.Host
		options.Token = s.Token
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Host, "sonarqube-host", c.Host, ""+
		"Sonarqube service address, if left empty, following sonarqube options will be ignored.")

	fs.StringVar(&s.Token, "sonarqube-token", c.Token, ""+
		"Sonarqube service access token.")
}
