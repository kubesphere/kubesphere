/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FakeS3 struct {
	Storage map[string]*Object
}

func NewFakeS3(objects ...*Object) *FakeS3 {
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

func (s *FakeS3) Read(key string) ([]byte, error) {
	if o, ok := s.Storage[key]; ok && o.Body != nil {
		data, err := ioutil.ReadAll(o.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, awserr.New(s3.ErrCodeNoSuchKey, "no such object", nil)
}
