package application

import (
	"context"
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/klog/v2"

	"io"

	"github.com/aws/aws-sdk-go/aws/awserr"
	s3lib "github.com/aws/aws-sdk-go/service/s3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv2 "kubesphere.io/api/application/v2"
	"kubesphere.io/utils/s3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type CmStore struct {
	Client runtimeclient.Client
}

var _ s3.Interface = CmStore{}

func InitStore(s3opts *s3.Options, Client client.Client) (cmStore, ossStore s3.Interface, err error) {
	if s3opts != nil && len(s3opts.Endpoint) != 0 {
		klog.Infof("init s3 client with endpoint: %s", s3opts.Endpoint)
		var err error
		ossStore, err = s3.NewS3Client(s3opts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create s3 client: %v", err)
		}
	}
	klog.Infof("init configmap store")
	cmStore = CmStore{
		Client: Client,
	}
	return cmStore, ossStore, nil
}

func (c CmStore) Read(key string) ([]byte, error) {
	cm := corev1.ConfigMap{}
	nameKey := runtimeclient.ObjectKey{Name: key, Namespace: appv2.ApplicationNamespace}
	err := c.Client.Get(context.TODO(), nameKey, &cm)
	if err != nil {
		return nil, err
	}
	return cm.BinaryData[appv2.BinaryKey], nil
}

func (c CmStore) Upload(key, fileName string, body io.Reader, size int) error {
	data, _ := io.ReadAll(body)
	obj := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key,
			Namespace: appv2.ApplicationNamespace,
		},
		BinaryData: map[string][]byte{appv2.BinaryKey: data},
	}
	err := c.Client.Create(context.TODO(), &obj)
	//ignore already exists error
	if apierrors.IsAlreadyExists(err) {
		klog.Warningf("save to store ignore already exists %s", key)
		return nil
	}

	return err
}

func (c CmStore) Delete(ids []string) error {
	for _, id := range ids {
		obj := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
				Namespace: appv2.ApplicationNamespace,
			},
		}
		err := c.Client.Delete(context.TODO(), &obj)
		if err != nil && apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func FailOverGet(cm, oss s3.Interface, key string, cli client.Client, isApp bool) (data []byte, err error) {

	if isApp {
		fromRepo, pullUrl, repoName, err := FromRepo(cli, key)
		if err != nil {
			klog.Errorf("failed to get app version, err: %v", err)
			return nil, err
		}
		if fromRepo {
			return DownLoadChart(cli, pullUrl, repoName)
		}
	}

	if oss == nil {
		klog.Infof("read from configMap %s", key)
		return cm.Read(key)
	}
	klog.Infof("read from oss %s", key)
	data, err = oss.Read(key)
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) && aerr.Code() == s3lib.ErrCodeNoSuchKey {
			klog.Infof("FailOver read from configMap %s", key)
			return cm.Read(key)
		}
	}
	return data, err
}

func FromRepo(cli runtimeclient.Client, key string) (fromRepo bool, url, repoName string, err error) {

	appVersion := appv2.ApplicationVersion{}
	err = cli.Get(context.TODO(), client.ObjectKey{Name: key}, &appVersion)
	if err != nil {
		klog.Errorf("failed to get app version, err: %v", err)
		return fromRepo, url, repoName, err
	}
	if appVersion.Spec.PullUrl != "" {
		klog.Infof("load chart from pull url: %s", appVersion.Spec.PullUrl)
	} else {
		klog.Infof("load chart from local store")
	}
	fromRepo = appVersion.GetLabels()[appv2.RepoIDLabelKey] != appv2.UploadRepoKey
	repoName = appVersion.GetLabels()[appv2.RepoIDLabelKey]
	return fromRepo, appVersion.Spec.PullUrl, repoName, nil
}

func DownLoadChart(cli runtimeclient.Client, pullUrl, repoName string) (data []byte, err error) {

	repo := appv2.Repo{}
	err = cli.Get(context.TODO(), client.ObjectKey{Name: repoName}, &repo)
	if err != nil {
		klog.Errorf("failed to get app repo, err: %v", err)
		return data, err
	}
	if registry.IsOCI(pullUrl) {
		return HelmPullFromOci(pullUrl, repo.Spec.Credential)
	}
	buf, err := HelmPull(pullUrl, repo.Spec.Credential)
	if err != nil {
		klog.Errorf("load chart failed, error: %s", err)
		return data, err
	}
	return buf.Bytes(), nil
}

func FailOverUpload(cm, oss s3.Interface, key string, body io.Reader, size int) error {
	if oss == nil {
		klog.Infof("upload to cm %s", key)
		return cm.Upload(key, key, body, size)
	}
	klog.Infof("upload to oss %s", key)
	return oss.Upload(key, key, body, size)
}
func FailOverDelete(cm, oss s3.Interface, key []string) error {
	if oss == nil {
		klog.Infof("delete from cm %v", key)
		return cm.Delete(key)
	}
	klog.Infof("delete from oss %v", key)
	return oss.Delete(key)
}
