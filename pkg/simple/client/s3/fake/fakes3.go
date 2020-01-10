package fake

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

type FakeS3 struct {
	storage map[string]*object
}

func NewFakeS3() *FakeS3 {
	return &FakeS3{storage: map[string]*object{}}
}

type object struct {
	key      string
	fileName string
	body     io.Reader
}

func (s *FakeS3) Upload(key, fileName string, body io.Reader) error {
	s.storage[key] = &object{
		key:      key,
		fileName: fileName,
		body:     body,
	}
	return nil
}

func (s *FakeS3) GetDownloadURL(key string, fileName string) (string, error) {
	if o, ok := s.storage[key]; ok {
		return fmt.Sprintf("http://%s/%s", o.key, fileName), nil
	}
	return "", awserr.New(s3.ErrCodeNoSuchKey, "no such object", nil)
}

func (s *FakeS3) Delete(key string) error {
	delete(s.storage, key)
	return nil
}
