package sonarqube

import (
	"github.com/spf13/pflag"
)

type SonarQubeOptions struct {
	Host  string `json:",omitempty" yaml:"host" description:"SonarQube service host address"`
	Token string `json:",omitempty" yaml:"token" description:"SonarQube service token"`
}

func NewSonarQubeOptions() *SonarQubeOptions {
	return &SonarQubeOptions{
		Host:  "",
		Token: "",
	}
}

func NewDefaultSonarQubeOptions() *SonarQubeOptions {
	return NewSonarQubeOptions()
}

func (s *SonarQubeOptions) Validate() []error {
	errors := []error{}

	return errors
}

func (s *SonarQubeOptions) ApplyTo(options *SonarQubeOptions) {
	if s.Host != "" {
		options.Host = s.Host
		options.Token = s.Token
	}
}

func (s *SonarQubeOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Host, "sonarqube-host", s.Host, ""+
		"Sonarqube service address, if left empty, following sonarqube options will be ignored.")

	fs.StringVar(&s.Token, "sonarqube-token", s.Token, ""+
		"Sonarqube service access token.")
}
