package s2is3

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
)

var (
	s3Region          string
	s3Endpoint        string
	s3DisableSSL      bool
	s3ForcePathStyle  bool
	s3AccessKeyID     string
	s3SecretAccessKey string
	s3SessionToken    string
	s3Bucket          string
)
var s2iS3 *s3.S3
var s2iS3Session *session.Session

func init() {
	flag.StringVar(&s3Region, "s2i-s3-region", "us-east-1", "region of s2i s3")
	flag.StringVar(&s3Endpoint, "s2i-s3-endpoint", "http://ks-minio.kubesphere-system.svc", "endpoint of s2i s3")
	flag.StringVar(&s3AccessKeyID, "s2i-s3-access-key-id", "AKIAIOSFODNN7EXAMPLE", "access key of s2i s3")
	flag.StringVar(&s3SecretAccessKey, "s2i-s3-secret-access-key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "secret access key of s2i s3")
	flag.StringVar(&s3SessionToken, "s2i-s3-session-token", "", "session token of s2i s3")
	flag.StringVar(&s3Bucket, "s2i-s3-bucket", "s2i-binaries", "bucket name of s2i s3")
	flag.BoolVar(&s3DisableSSL, "s2i-s3-disable-SSL", true, "disable ssl")
	flag.BoolVar(&s3ForcePathStyle, "s2i-s3-force-path-style", true, "force path style")
}

func Client() *s3.S3 {
	if s2iS3 != nil {
		return s2iS3
	}
	creds := credentials.NewStaticCredentials(
		s3AccessKeyID, s3SecretAccessKey, s3SessionToken,
	)
	config := &aws.Config{
		Region:           aws.String(s3Region),
		Endpoint:         aws.String(s3Endpoint),
		DisableSSL:       aws.Bool(s3DisableSSL),
		S3ForcePathStyle: aws.Bool(s3ForcePathStyle),
		Credentials:      creds,
	}
	sess, err := session.NewSession(config)
	if err != nil {
		glog.Errorf("failed to connect to s2i s3: %+v", err)
		return nil
	}
	s2iS3Session = sess
	s2iS3 = s3.New(sess)
	return s2iS3
}
func Session() *session.Session {
	if s2iS3Session != nil {
		return s2iS3Session
	}
	Client()
	return s2iS3Session
}

func Bucket() *string {
	return aws.String(s3Bucket)
}
