package s2is3

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

// S3Options contains configuration to access a s3 service
type S3Options struct {
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint"`
	Region          string `json:"region,omitempty" yaml:"region"`
	DisableSSL      bool   `json:"disableSSL" yaml:"disableSSL"`
	ForcePathStyle  bool   `json:"forcePathStyle" yaml:"forcePathStyle"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey"`
	SessionToken    string `json:"sessionToken,omitempty" yaml:"sessionToken"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket"`
}

// NewS3Options creates a default disabled S3Options(empty endpoint)
func NewS3Options() *S3Options {
	return &S3Options{
		Endpoint:        "",
		Region:          "us-east-1",
		DisableSSL:      true,
		ForcePathStyle:  true,
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		SessionToken:    "",
		Bucket:          "s2i-binaries",
	}
}

// Validate check options values
func (s *S3Options) Validate() []error {
	errors := []error{}

	return errors
}

// ApplyTo overrides options if it's valid, which endpoint is not empty
func (s *S3Options) ApplyTo(options *S3Options) {
	if s.Endpoint != "" {
		reflectutils.Override(options, s)
	}
}

// AddFlags add options flags to command line flags,
// if s3-endpoint if left empty, following options will be ignored
func (s *S3Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Endpoint, "s3-endpoint", s.Endpoint, ""+
		"Endpoint to access to s3 object storage service, if left blank, the following options "+
		"will be ignored.")

	fs.StringVar(&s.Region, "s3-region", s.Region, ""+
		"Region of s3 that will access to, like us-east-1.")

	fs.StringVar(&s.AccessKeyID, "s3-access-key-id", s.AccessKeyID, "access key of s2i s3")

	fs.StringVar(&s.SecretAccessKey, "s3-secret-access-key", s.SecretAccessKey, "secret access key of s2i s3")

	fs.StringVar(&s.SessionToken, "s3-session-token", s.SessionToken, "session token of s2i s3")

	fs.StringVar(&s.Bucket, "s3-bucket", s.Bucket, "bucket name of s2i s3")

	fs.BoolVar(&s.DisableSSL, "s3-disable-SSL", s.DisableSSL, "disable ssl")

	fs.BoolVar(&s.ForcePathStyle, "s3-force-path-style", s.ForcePathStyle, "force path style")
}
