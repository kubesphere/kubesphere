package devops

import (
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"mime/multipart"
	"net/http"
	"reflect"
)

const (
	GetS2iBinaryURL = "http://ks-apiserver.kubesphere-system.svc/kapis/devops.kubesphere.io/v1alpha2/namespaces/%s/s2ibinaries/%s/file/%s"
)

type S2iBinaryUploader interface {
	UploadBinary(name, namespace, md5 string, header *multipart.FileHeader) (*v1alpha1.S2iBinary, error)

	DownloadBinary(name, namespace, fileName string) (string, error)
}

type s2iBinaryUploader struct {
	client    versioned.Interface
	informers externalversions.SharedInformerFactory
	s3Client  s3.Interface
}

func (s *s2iBinaryUploader) UploadS2iBinary(namespace, name, md5 string, fileHeader *multipart.FileHeader) (*v1alpha1.S2iBinary, error) {
	binFile, err := fileHeader.Open()
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, err
	}
	defer binFile.Close()

	origin, err := s.informers.Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(namespace).Get(name)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, err
	}
	//Check file is uploading
	if origin.Status.Phase == v1alpha1.StatusUploading {
		err := restful.NewError(http.StatusConflict, "file is uploading, please try later")
		klog.Error(err)
		return nil, err
	}
	copy := origin.DeepCopy()
	copy.Spec.MD5 = md5
	copy.Spec.Size = bytefmt.ByteSize(uint64(fileHeader.Size))
	copy.Spec.FileName = fileHeader.Filename
	copy.Spec.DownloadURL = fmt.Sprintf(GetS2iBinaryURL, namespace, name, copy.Spec.FileName)
	if origin.Status.Phase == v1alpha1.StatusReady && reflect.DeepEqual(origin, copy) {
		return origin, nil
	}

	//Set status Uploading to lock resource
	uploading, err := s.SetS2iBinaryStatus(copy, v1alpha1.StatusUploading)
	if err != nil {
		err := restful.NewError(http.StatusConflict, fmt.Sprintf("could not set status: %+v", err))
		klog.Error(err)
		return nil, err
	}

	copy = uploading.DeepCopy()
	copy.Spec.MD5 = md5
	copy.Spec.Size = bytefmt.ByteSize(uint64(fileHeader.Size))
	copy.Spec.FileName = fileHeader.Filename
	copy.Spec.DownloadURL = fmt.Sprintf(GetS2iBinaryURL, namespace, name, copy.Spec.FileName)

	/*
		TODO: upload binary file use s3.Interface
		s3session := s3Client.Session()
		if s3session == nil {
			err := fmt.Errorf("could not connect to s2i s3")
			klog.Error(err)
			_, serr := SetS2iBinaryStatusWithRetry(copy, origin.Status.Phase)
			if serr != nil {
				klog.Error(serr)
				return nil, err
			}
			return nil, err
		}
		uploader := s3manager.NewUploader(s3session, func(uploader *s3manager.Uploader) {
			uploader.PartSize = 5 * bytefmt.MEGABYTE
			uploader.LeavePartsOnError = true
		})
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:             s3Client.Bucket(),
			Key:                aws.String(fmt.Sprintf("%s-%s", namespace, name)),
			Body:               binFile,
			ContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", copy.Spec.FileName)),
		})

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					klog.Error(err)
					_, serr := SetS2iBinaryStatusWithRetry(copy, origin.Status.Phase)
					if serr != nil {
						klog.Error(serr)
					}
					return nil, err
				default:
					klog.Error(err)
					_, serr := SetS2iBinaryStatusWithRetry(copy, v1alpha1.StatusUploadFailed)
					if serr != nil {
						klog.Error(serr)
					}
					return nil, err
				}
			}
			klog.Error(err)
			return nil, err
		}
	*/
	if copy.Spec.UploadTimeStamp == nil {
		copy.Spec.UploadTimeStamp = new(metav1.Time)
	}
	*copy.Spec.UploadTimeStamp = metav1.Now()
	copy, err = client.ClientSets().K8s().KubeSphere().DevopsV1alpha1().S2iBinaries(namespace).Update(copy)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	copy, err = s.SetS2iBinaryStatusWithRetry(copy, v1alpha1.StatusReady)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return copy, nil
}

func (s *s2iBinaryUploader) DownloadS2iBinary(namespace, name, fileName string) (string, error) {

	origin, err := s.informers.Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(namespace).Get(name)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", err
	}

	if origin.Spec.FileName != fileName {
		err := fmt.Errorf("could not fould file %s", fileName)
		klog.Error(err)
		return "", err
	}

	if origin.Status.Phase != v1alpha1.StatusReady {
		err := restful.NewError(http.StatusBadRequest, "file is not ready, please try later")
		klog.Error(err)
		return "", err
	}

	/*
		TODO: get download url of requested binary
		req, _ := s.s3Client.Client().GetObjectRequest(&s3.GetObjectInput{
			Bucket:                     s3Client.Bucket(),
			Key:                        aws.String(fmt.Sprintf("%s-%s", namespace, name)),
			ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", origin.Spec.FileName)),
		})
		url, err := req.Presign(5 * time.Minute)
		if err != nil {
			klog.Error(err)
			return "", err
		}
	*/
	return "", nil

}

func (s *s2iBinaryUploader) SetS2iBinaryStatus(s2ibin *v1alpha1.S2iBinary, status string) (*v1alpha1.S2iBinary, error) {
	copy := s2ibin.DeepCopy()
	copy.Status.Phase = status
	copy, err := s.client.DevopsV1alpha1().S2iBinaries(s2ibin.Namespace).Update(copy)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return copy, nil
}

func (s *s2iBinaryUploader) SetS2iBinaryStatusWithRetry(s2ibin *v1alpha1.S2iBinary, status string) (*v1alpha1.S2iBinary, error) {

	var bin *v1alpha1.S2iBinary
	var err error
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		bin, err = s.informers.Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(s2ibin.Namespace).Get(s2ibin.Name)
		if err != nil {
			klog.Error(err)
			return err
		}
		bin.Status.Phase = status
		bin, err = s.client.DevopsV1alpha1().S2iBinaries(s2ibin.Namespace).Update(bin)
		if err != nil {
			klog.Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return bin, nil
}
