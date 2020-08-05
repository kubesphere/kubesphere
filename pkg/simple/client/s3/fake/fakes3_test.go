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
	"testing"
)

func TestFakeS3(t *testing.T) {
	s3 := NewFakeS3()
	key := "hello"
	fileName := "world"
	err := s3.Upload(key, fileName, nil)
	if err != nil {
		t.Fatal(err)
	}
	o, ok := s3.Storage["hello"]
	if !ok {
		t.Fatal("should have hello object")
	}
	if o.Key != key || o.FileName != fileName {
		t.Fatalf("Key should be %s, FileName should be %s", key, fileName)
	}

	url, err := s3.GetDownloadURL(key, fileName+"1")
	if err != nil {
		t.Fatal(err)
	}
	if url != fmt.Sprintf("http://%s/%s", key, fileName+"1") {
		t.Fatalf("url should be %s", fmt.Sprintf("http://%s/%s", key, fileName+"1"))
	}

	url, err = s3.GetDownloadURL(key, fileName+"2")
	if err != nil {
		t.Fatal(err)
	}
	if url != fmt.Sprintf("http://%s/%s", key, fileName+"2") {
		t.Fatalf("url should be %s", fmt.Sprintf("http://%s/%s", key, fileName+"2"))
	}

	err = s3.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
	_, ok = s3.Storage["hello"]
	if ok {
		t.Fatal("should not have hello object")
	}
	err = s3.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
}
