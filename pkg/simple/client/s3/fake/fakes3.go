package fake

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	clients3 "kubesphere.io/kubesphere/pkg/simple/client/s3"
)

type FakeS3 struct {
	Storage map[string]*Object
}

func NewFakeS3(objects ...*Object) clients3.Interface {
	s3 := &FakeS3{Storage: map[string]*Object{}}
	for _, object := range objects {
		s3.Storage[object.Key] = object
	}
	return s3
}

type Object struct {
	Key      string
	FileName string
	Body     io.Reader
}

func (s *FakeS3) Upload(key, fileName string, body io.Reader) error {
	s.Storage[key] = &Object{
		Key:      key,
		FileName: fileName,
		Body:     body,
	}
	return nil
}

func (s *FakeS3) GetDownloadURL(key string, fileName string) (string, error) {
	if o, ok := s.Storage[key]; ok {
		return fmt.Sprintf("http://%s/%s", o.Key, fileName), nil
	}
	return "", awserr.New(s3.ErrCodeNoSuchKey, "no such object", nil)
}

func (s *FakeS3) Delete(key string) error {
	delete(s.Storage, key)
	return nil
}
