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
	o, ok := s3.storage["hello"]
	if !ok {
		t.Fatal("should have hello object")
	}
	if o.key != key || o.fileName != fileName {
		t.Fatalf("key should be %s, fileName should be %s", key, fileName)
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
	_, ok = s3.storage["hello"]
	if ok {
		t.Fatal("should not have hello object")
	}
	err = s3.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
}
