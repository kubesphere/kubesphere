package s2is3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s.io/klog"
)

type S3Client struct {
	s3Client  *s3.S3
	s3Session *session.Session
	bucket    string
}

func NewS3Client(options *S3Options) (*S3Client, error) {
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

	var c S3Client

	c.s3Client = s3.New(s)
	c.s3Session = s
	c.bucket = options.Bucket

	return &c, nil
}

// NewS3ClientOrDie creates S3Client and panics if there is an error
func NewS3ClientOrDie(options *S3Options) *S3Client {
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

	return &S3Client{
		s3Client:  client,
		s3Session: s,
		bucket:    options.Bucket,
	}
}

func (s *S3Client) Client() *s3.S3 {

	return s.s3Client
}
func (s *S3Client) Session() *session.Session {
	return s.s3Session
}

func (s *S3Client) Bucket() *string {
	return aws.String(s.bucket)
}
