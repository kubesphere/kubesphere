package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"k8s.io/klog"
	"time"
)

type Client struct {
	s3Client  *s3.S3
	s3Session *session.Session
	bucket    string
}

func (s *Client) Upload(key string, body io.Reader) (string, error) {
	panic("implement me")
}

func (s *Client) Get(key string, fileName string, expire time.Duration) (string, error) {
	panic("implement me")
}

func (s *Client) Delete(key string) error {
	panic("implement me")
}

func NewS3Client(options *Options) (Interface, error) {
	cred := credentials.NewStaticCredentials(options.AccessKeyID, options.SecretAccessKey, options.SessionToken)

	config := aws.Config{
		Region:           aws.String(options.Region),
		Endpoint:         aws.String(options.Endpoint),
		DisableSSL:       aws.Bool(options.DisableSSL),
		S3ForcePathStyle: aws.Bool(options.ForcePathStyle),
		Credentials:      cred,
	}

	s, err := session.NewSession(&config)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var c Client

	c.s3Client = s3.New(s)
	c.s3Session = s
	c.bucket = options.Bucket

	return &c, nil
}

// NewS3ClientOrDie creates Client and panics if there is an error
func NewS3ClientOrDie(options *Options) Interface {
	cred := credentials.NewStaticCredentials(options.AccessKeyID, options.SecretAccessKey, options.SessionToken)

	config := aws.Config{
		Region:           aws.String(options.Region),
		Endpoint:         aws.String(options.Endpoint),
		DisableSSL:       aws.Bool(options.DisableSSL),
		S3ForcePathStyle: aws.Bool(options.ForcePathStyle),
		Credentials:      cred,
	}

	s, err := session.NewSession(&config)
	if err != nil {
		panic(err)
	}

	client := s3.New(s)

	return &Client{
		s3Client:  client,
		s3Session: s,
		bucket:    options.Bucket,
	}
}

func (s *Client) Client() *s3.S3 {

	return s.s3Client
}
func (s *Client) Session() *session.Session {
	return s.s3Session
}

func (s *Client) Bucket() *string {
	return aws.String(s.bucket)
}
