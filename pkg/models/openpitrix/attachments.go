///*
//Copyright 2020 The KubeSphere Authors.
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package openpitrix

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/afero/mem"
	"io"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	informers2 "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"sync"
)

var attachmentHandler attachmentCtl = &AttachmentCtl{}

const (
	AttachmentSecretName    = "kubesphere-secret"
	AttachmentSecretKeyName = "attachment_config"
)

type AttachmentInterface interface {
	DescribeAttachment(id string) (*Attachment, error)
}

type attachmentOperator struct{}

func newAttachmentOperator(k8sFactory informers2.SharedInformerFactory) AttachmentInterface {
	if k8sFactory != nil {
		k8sFactory.Core().V1().Secrets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				if secret.Namespace == constants.KubeSphereNamespace && secret.Name == AttachmentSecretName {
					att := attachmentHandler.(*AttachmentCtl)
					att.LoadConfig(secret)
				}
			},
			DeleteFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				if secret.Namespace == constants.KubeSphereNamespace && secret.Name == AttachmentSecretName {
					att := attachmentHandler.(*AttachmentCtl)
					att.Lock()
					att.attachmentCtl = nil
					att.Unlock()
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				secret := newObj.(*v1.Secret)
				if secret.Namespace == constants.KubeSphereNamespace && secret.Name == AttachmentSecretName {
					att := attachmentHandler.(*AttachmentCtl)
					att.LoadConfig(secret)
				}
			},
		})
	}

	return &attachmentOperator{}
}

func (c *attachmentOperator) DescribeAttachment(id string) (*Attachment, error) {
	r, err := attachmentHandler.Read(id)

	if err != nil {
		return nil, err
	}
	att := &Attachment{AttachmentID: id}
	if r == nil {
		return att, nil
	}
	b, err := ioutil.ReadAll(r)
	r.Close()

	if err != nil {
		return nil, err
	} else {
		att.AttachmentContent = map[string]strfmt.Base64{
			"raw": b,
		}
	}

	return att, nil
}

type AttachmentCtl struct {
	attachmentCtl
	sync.RWMutex
}

func (a *AttachmentCtl) LoadConfig(secret *v1.Secret) {
	if secret == nil {
		return
	}
	klog.V(2).Infof("load attachment_config from secret")
	data := secret.Data[AttachmentSecretKeyName]
	if len(data) == 0 {
		a.Lock()
		if a.attachmentCtl == nil {
			a.attachmentCtl = nil
			klog.Warningf("remove attachment_config from secret")
		}
		a.RLock()
		return
	} else {
		attachConfig := &AttachmentConfig{}
		err := json.Unmarshal(data, attachConfig)
		if err != nil {
			klog.Errorf("json decode attachment_config from secret failed, error: %s", err)
			return
		}

		if attachConfig.S3Config != nil {
			a := attachmentHandler.(*AttachmentCtl)
			a.Lock()
			a.attachmentCtl = attachConfig.S3Config
			a.Unlock()
			klog.Infof("load attachment_config from secret success")
		}
	}

	return
}

func (a *AttachmentCtl) Read(id string) (io.ReadCloser, error) {
	a.RLock()
	defer a.RUnlock()
	if a.attachmentCtl == nil {
		klog.V(2).Infof("uninitialized attachment_config")
		return nil, nil
	}
	return a.attachmentCtl.Read(id)
}

func (a *AttachmentCtl) Delete(id string) error {
	a.RLock()
	defer a.RUnlock()
	if a.attachmentCtl == nil {
		klog.V(2).Infof("uninitialized attachment_config")
		return nil
	}
	return a.attachmentCtl.Delete(id)
}

func (a *AttachmentCtl) Save(id string, r io.Reader) error {
	a.RLock()
	defer a.RUnlock()
	if a.attachmentCtl == nil {
		klog.V(2).Infof("uninitialized attachment_config")
		return nil
	}
	return a.attachmentCtl.Save(id, r)
}

type attachmentCtl interface {
	//get attachment id
	Read(id string) (io.ReadCloser, error)
	//save attachment as name `id`
	Save(id string, reader io.Reader) error
	//delete attachment id
	Delete(id string) error
}

type AttachmentConfig struct {
	S3Config *S3Config `json:"s3_config,omitempty"`
}

type S3Config struct {
	AccessKey  string `json:"access_key" yaml:"access_key"`
	SecretKey  string `json:"secret_key" yaml:"secret_key"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Bucket     string `json:"bucket" yaml:"bucket"`
	Region     string `json:"region" yaml:"region"`
	DisableSSL bool   `json:"disable_ssl" yaml:"disable_ssl"`
}

func (s *S3Config) newSession() (*session.Session, error) {
	region := s.Region
	if region == "" {
		region = "us-east-1"
	}
	return session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(s.AccessKey, s.SecretKey, ""),
		Endpoint:         aws.String(s.Endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
}

func (s *S3Config) Delete(id string) error {
	sess, err := s.newSession()
	svc := s3.New(sess)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(id)})
	if err != nil {
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(id),
	})

	return err
}

func (s *S3Config) Save(id string, reader io.Reader) error {
	sess, err := s.newSession()
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(id),
		Body:   reader,
	})

	return err
}

func (s *S3Config) Read(id string) (io.ReadCloser, error) {

	sess, err := s.newSession()
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(sess)

	f := mem.NewFileHandle(mem.CreateFile(id))
	f.Open()
	_, err = downloader.Download(f,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(id),
		})

	if err != nil {
		return nil, err
	}

	f.Seek(0, io.SeekStart)

	return f, nil
}
