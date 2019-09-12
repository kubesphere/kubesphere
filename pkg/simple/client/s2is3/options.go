package s2is3

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type S3Options struct {
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty"`
	DisableSSL      bool   `json:"disableSSL,omitempty" yaml:"disableSSL,omitempty"`
	ForcePathStyle  bool   `json:"forcePathStyle,omitempty" yaml:"forePathStyle,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty" yaml:"sessionToken,omitempty"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
}

func NewS3Options() *S3Options {
	return &S3Options{
		Endpoint:        "",
		Region:          "",
		DisableSSL:      true,
		ForcePathStyle:  true,
		AccessKeyID:     "",
		SecretAccessKey: "",
		SessionToken:    "",
		Bucket:          "",
	}
}

func (s *S3Options) Validate() []error {
	errors := []error{}

	return errors
}

func (s *S3Options) ApplyTo(options *S3Options) {
	reflectutils.Override(options, s)
}

func (s *S3Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Endpoint, "s3-endpoint", s.Endpoint, ""+
		"Endpoint to access to s3 object storage service, if left blank, the following options "+
		"will be ignored.")

	fs.StringVar(&s.Region, "s3-region", s.Region, ""+
		"Region of s3 that will access to, like us-east-1.")

	fs.StringVar(&s.AccessKeyID, "s3-access-key-id", "AKIAIOSFODNN7EXAMPLE", "access key of s2i s3")

	fs.StringVar(&s.SecretAccessKey, "s3-secret-access-key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "secret access key of s2i s3")

	fs.StringVar(&s.SessionToken, "s3-session-token", s.SessionToken, "session token of s2i s3")

	fs.StringVar(&s.Bucket, "s3-bucket", "s2i-binaries", "bucket name of s2i s3")

	fs.BoolVar(&s.DisableSSL, "s3-disable-SSL", s.DisableSSL, "disable ssl")

	fs.BoolVar(&s.ForcePathStyle, "s3-force-path-style", true, "force path style")
}
