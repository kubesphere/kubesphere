package openpitrix

import (
	"fmt"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type OpenPitrixOptions struct {
	RuntimeManagerEndpoint    string `json:"runtimeManagerEndpoint,omitempty" yaml:"runtimeManagerEndpoint,omitempty"`
	ClusterManagerEndpoint    string `json:"clusterManagerEndpoint,omitempty" yaml:"clusterManagerEndpoint,omitempty"`
	RepoManagerEndpoint       string `json:"repoManagerEndpoint,omitempty" yaml:"repoManagerEndpoint,omitempty"`
	AppManagerEndpoint        string `json:"appManagerEndpoint,omitempty" yaml:"appManagerEndpoint,omitempty"`
	CategoryManagerEndpoint   string `json:"categoryManagerEndpoint,omitempty" yaml:"categoryManagerEndpoint,omitempty"`
	AttachmentManagerEndpoint string `json:"attachmentManagerEndpoint,omitempty" yaml:"attachmentManagerEndpoint,omitempty"`
	RepoIndexerEndpoint       string `json:"repoIndexerEndpoint,omitempty" yaml:"repoIndexerEndpoint,omitempty"`
}

func NewOpenPitrixOptions() *OpenPitrixOptions {
	return &OpenPitrixOptions{}
}

func (s *OpenPitrixOptions) ApplyTo(options *OpenPitrixOptions) {
	if options == nil {
		options = s
		return
	}
	if s.RuntimeManagerEndpoint != "" {
		reflectutils.Override(options, s)
	}
}

func (s *OpenPitrixOptions) IsEmpty() bool {
	return s.RuntimeManagerEndpoint == "" &&
		s.ClusterManagerEndpoint == "" &&
		s.RepoManagerEndpoint == "" &&
		s.AppManagerEndpoint == "" &&
		s.CategoryManagerEndpoint == "" &&
		s.AttachmentManagerEndpoint == "" &&
		s.RepoIndexerEndpoint == ""
}

func (s *OpenPitrixOptions) Validate() []error {
	var errs []error

	if s.RuntimeManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.RuntimeManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.RuntimeManagerEndpoint))
		}
	}
	if s.ClusterManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.ClusterManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.ClusterManagerEndpoint))
		}
	}
	if s.RepoManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.RepoManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.RepoManagerEndpoint))
		}
	}
	if s.RepoIndexerEndpoint != "" {
		_, _, err := parseToHostPort(s.RepoIndexerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.RepoIndexerEndpoint))
		}
	}
	if s.AppManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.AppManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.AppManagerEndpoint))
		}
	}
	if s.CategoryManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.CategoryManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.CategoryManagerEndpoint))
		}
	}
	if s.AttachmentManagerEndpoint != "" {
		_, _, err := parseToHostPort(s.CategoryManagerEndpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid host port:%s", s.CategoryManagerEndpoint))
		}
	}

	return errs
}

func (s *OpenPitrixOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.RuntimeManagerEndpoint, "openpitrix-runtime-manager-endpoint", s.RuntimeManagerEndpoint, ""+
		"OpenPitrix runtime manager endpoint")

	fs.StringVar(&s.AppManagerEndpoint, "openpitrix-app-manager-endpoint", s.AppManagerEndpoint, ""+
		"OpenPitrix app manager endpoint")

	fs.StringVar(&s.ClusterManagerEndpoint, "openpitrix-cluster-manager-endpoint", s.ClusterManagerEndpoint, ""+
		"OpenPitrix cluster manager endpoint")

	fs.StringVar(&s.CategoryManagerEndpoint, "openpitrix-category-manager-endpoint", s.CategoryManagerEndpoint, ""+
		"OpenPitrix category manager endpoint")

	fs.StringVar(&s.RepoManagerEndpoint, "openpitrix-repo-manager-endpoint", s.RepoManagerEndpoint, ""+
		"OpenPitrix repo manager endpoint")

	fs.StringVar(&s.RepoIndexerEndpoint, "openpitrix-repo-indexer-endpoint", s.RepoIndexerEndpoint, ""+
		"OpenPitrix repo indexer endpoint")

	fs.StringVar(&s.AttachmentManagerEndpoint, "openpitrix-attachment-manager-endpoint", s.AttachmentManagerEndpoint, ""+
		"OpenPitrix attachment manager endpoint")
}
