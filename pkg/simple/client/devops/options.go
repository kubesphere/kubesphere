package devops

import (
	"fmt"
	"github.com/spf13/pflag"
)

type DevopsOptions struct {
	Host           string `json:",omitempty" yaml:",omitempty" description:"Jenkins service host address"`
	Username       string `json:",omitempty" yaml:",omitempty" description:"Jenkins admin username"`
	Password       string `json:",omitempty" yaml:",omitempty" description:"Jenkins admin password"`
	MaxConnections int    `json:"maxConnections,omitempty" yaml:"maxConnections,omitempty" description:"Maximum connections allowed to connect to Jenkins"`
}

// NewDevopsOptions returns a `zero` instance
func NewDevopsOptions() *DevopsOptions {
	return &DevopsOptions{
		Host:           "",
		Username:       "",
		Password:       "",
		MaxConnections: 100,
	}
}

func (s *DevopsOptions) ApplyTo(options *DevopsOptions) {
	if options == nil {
		return
	}

	if s.Host != "" {
		options.Host = s.Host
	}

	if s.Username != "" {
		options.Username = s.Username
	}

	if s.Password != "" {
		options.Password = s.Password
	}

	if s.MaxConnections > 0 {
		options.MaxConnections = s.MaxConnections
	}
}

//
func (s *DevopsOptions) Validate() []error {
	errors := []error{}

	// devops is not needed, ignore rest options
	if s.Host == "" {
		return errors
	}

	if s.Username == "" || s.Password == "" {
		errors = append(errors, fmt.Errorf("jenkins's username or password is empty"))
	}

	if s.MaxConnections <= 0 {
		errors = append(errors, fmt.Errorf("jenkins's maximum connections should be greater than 0"))
	}

	return errors
}

func (s *DevopsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Host, "jenkins-host", s.Host, ""+
		"Jenkins service host address. If left blank, means Jenkins "+
		"is unnecessary.")

	fs.StringVar(&s.Username, "jenkins-username", s.Username, ""+
		"Username for access to Jenkins service. Leave it blank if there isn't any.")

	fs.StringVar(&s.Password, "jenkins-password", s.Password, ""+
		"Password for access to Jenkins service, used pair with username.")

	fs.IntVar(&s.MaxConnections, "jenkins-max-connections", s.MaxConnections, ""+
		"Maximum allowed connections to Jenkins. ")

}
