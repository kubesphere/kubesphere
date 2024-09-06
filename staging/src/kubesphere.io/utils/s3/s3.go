package s3

import (
	"fmt"
	"io"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"k8s.io/klog/v2"
)

type Interface interface {
	Read(key string) ([]byte, error)
	Upload(key, fileName string, body io.Reader, size int) error
	Delete(key []string) error
}

type Client struct {
	s3Client  *s3.S3
	s3Session *session.Session
	bucket    string
}

// Options contains configuration to access a s3 service
type Options struct {
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty"`
	DisableSSL      bool   `json:"disableSSL" yaml:"disableSSL"`
	ForcePathStyle  bool   `json:"forcePathStyle" yaml:"forcePathStyle"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty" yaml:"sessionToken,omitempty"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
}

// NewS3Options creates a default disabled Options(empty endpoint)
func NewS3Options() *Options {
	return &Options{
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

const (
	DefaultPartSize = 5 << (10 * 2)
	// MinConcurrency is the minimum concurrency when uploading a part to Amazon S3,
	// it's also the default value of Concurrency in aws-sdk-go.
	MinConcurrency = 5
	// MaxConcurrency is the maximum concurrency to limit the goroutines.
	MaxConcurrency = 128
)

// calculateConcurrency calculates the concurrency for better performance,
// make the concurrency in range [5, 128].
func calculateConcurrency(size int) int {
	if size <= 0 {
		return MinConcurrency
	}
	c := int(math.Ceil(float64(size) / float64(DefaultPartSize)))
	if c < MinConcurrency {
		return MinConcurrency
	} else if c > MaxConcurrency {
		return MaxConcurrency
	}
	return c
}

// Upload use Multipart upload to upload a single object as a set of parts.
// If the data length is known to be large, it is recommended to pass in the data length,
// it will help to calculate concurrency. Otherwise, `size` can be 0,
// use 5 as default upload concurrency, same as aws-sdk-go.
// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html for more details.
func (s *Client) Upload(key, fileName string, body io.Reader, size int) error {
	uploader := s3manager.NewUploader(s.s3Session, func(uploader *s3manager.Uploader) {
		uploader.PartSize = DefaultPartSize
		uploader.LeavePartsOnError = true
		uploader.Concurrency = calculateConcurrency(size)
	})
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(s.bucket),
		Key:                aws.String(key),
		Body:               body,
		ContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", fileName)),
	})
	return err
}

func (s *Client) Read(key string) ([]byte, error) {

	downloader := s3manager.NewDownloader(s.s3Session)

	writer := aws.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(writer,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

func (s *Client) Delete(ids []string) error {
	for _, key := range ids {
		_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
		if err != nil {
			klog.Errorf("failed to delete object %s, err: %v", key, err)
			return err
		}
	}
	return nil
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

func (s *Client) Client() *s3.S3 {
	return s.s3Client
}
func (s *Client) Session() *session.Session {
	return s.s3Session
}

func (s *Client) Bucket() *string {
	return aws.String(s.bucket)
}
