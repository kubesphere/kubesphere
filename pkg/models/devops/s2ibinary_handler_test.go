/*
Copyright 2020 The KubeSphere Authors.

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

package devops

import (
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientgotesting "k8s.io/client-go/testing"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	fakeS3 "kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"kubesphere.io/kubesphere/pkg/utils/hashutil"
	"mime/multipart"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	fileaContents = "This is a test file."
	fileaKey      = "binary"
	fileaName     = "filea.txt"
	boundary      = `MyBoundary`
	ns            = "testns"
	s2ibname      = "test"
)

const message = `
--MyBoundary
Content-Disposition: form-data; name="binary"; filename="filea.txt"
Content-Type: text/plain

` + fileaContents + `
--MyBoundary--
`

func TestS2iBinaryUploader(t *testing.T) {
	s2ib := s2ibinary(ns, s2ibname)
	fakeKubeClient := fake.NewSimpleClientset(s2ib)
	fakeWatch := watch.NewFake()
	fakeKubeClient.AddWatchReactor("*", clientgotesting.DefaultWatchReactor(fakeWatch, nil))
	informerFactory := ksinformers.NewSharedInformerFactory(fakeKubeClient, 0)
	stopCh := make(chan struct{})
	s2iInformer := informerFactory.Devops().V1alpha1().S2iBinaries()
	err := s2iInformer.Informer().GetIndexer().Add(s2ib)
	defer close(stopCh)
	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	s3 := fakeS3.NewFakeS3()
	uploader := NewS2iBinaryUploader(fakeKubeClient, informerFactory, s3)
	header := prepareFileHeader()
	file, err := header.Open()
	if err != nil {
		t.Fatal(err)
	}
	md5, err := hashutil.GetMD5(file)
	if err != nil {
		t.Fatal(err)
	}
	wantSpec := v1alpha1.S2iBinarySpec{
		FileName: fileaName,
		MD5:      md5,
		Size:     bytefmt.ByteSize(uint64(header.Size)),
	}

	binary, err := uploader.UploadS2iBinary(ns, s2ibname, md5, header)
	if err != nil {
		t.Fatal(err)
	}

	wantSpec.UploadTimeStamp = binary.Spec.UploadTimeStamp
	wantSpec.DownloadURL = binary.Spec.DownloadURL
	if !reflect.DeepEqual(binary.Spec, wantSpec) {
		t.Fatalf("s2ibinary spec is not same with expected, get: %+v, expected: %+v", binary, wantSpec)
	}

	_, ok := s3.Storage[fmt.Sprintf("%s-%s", ns, s2ibname)]
	if !ok {
		t.Fatalf("should get file in s3")
	}

	time.Sleep(3 * time.Second)
	url, err := uploader.DownloadS2iBinary(ns, s2ibname, fileaName)
	if err != nil {
		t.Fatal(err)
	}
	if url != fmt.Sprintf("http://%s-%s/%s", ns, s2ibname, fileaName) {
		t.Fatalf("download url is not equal with expected, get: %+v, expected: %+v", url, fmt.Sprintf("http://%s-%s/%s", ns, s2ibname, fileaName))
	}
}

func prepareFileHeader() *multipart.FileHeader {
	reader := strings.NewReader(message)
	multipartReader := multipart.NewReader(reader, boundary)
	form, err := multipartReader.ReadForm(25)
	if err != nil {
		panic(err)
	}
	return form.File["binary"][0]
}

func s2ibinary(namespace, name string) *v1alpha1.S2iBinary {
	return &v1alpha1.S2iBinary{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec:   v1alpha1.S2iBinarySpec{},
		Status: v1alpha1.S2iBinaryStatus{},
	}
}
