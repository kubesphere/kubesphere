package devops

import (
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"mime/multipart"
	"net/http"
	"reflect"
	"time"
)

const (
	GetS2iBinaryURL = "http://ks-apiserver.kubesphere-system.svc/kapis/devops.kubesphere.io/v1alpha2/namespaces/%s/s2ibinaries/%s/file/%s"
)

func UploadS2iBinary(namespace, name, md5 string, fileHeader *multipart.FileHeader) (*v1alpha1.S2iBinary, error) {
	binFile, err := fileHeader.Open()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err
	}
	defer binFile.Close()

	origin, err := informers.KsSharedInformerFactory().Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(namespace).Get(name)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err
	}
	//Check file is uploading
	if origin.Status.Phase == v1alpha1.StatusUploading {
		err := restful.NewError(http.StatusConflict, "file is uploading, please try later")
		glog.Error(err)
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
	origin, err = SetS2iBinaryStatus(origin, v1alpha1.StatusUploading)
	if err != nil {
		err := restful.NewError(http.StatusConflict, fmt.Sprintf("could not set status: %+v", err))
		glog.Error(err)
		return nil, err
	}
	copy = origin.DeepCopy()
	copy.Spec.MD5 = md5
	copy.Spec.Size = bytefmt.ByteSize(uint64(fileHeader.Size))
	copy.Spec.FileName = fileHeader.Filename
	copy.Spec.DownloadURL = fmt.Sprintf(GetS2iBinaryURL, namespace, name, copy.Spec.FileName)
	s3session := s2is3.Session()
	if s3session == nil {
		err := fmt.Errorf("could not connect to s2i s3")
		glog.Error(err)
		return nil, err
	}
	uploader := s3manager.NewUploader(s3session, func(uploader *s3manager.Uploader) {
		uploader.PartSize = 5 * bytefmt.MEGABYTE
		uploader.LeavePartsOnError = true
	})
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:             s2is3.Bucket(),
		Key:                aws.String(fmt.Sprintf("%s-%s", namespace, name)),
		Body:               binFile,
		ContentMD5:         aws.String(md5),
		ContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", copy.Spec.FileName)),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				glog.Error(err)
				_, serr := SetS2iBinaryStatusWithRetry(origin, origin.Status.Phase)
				if serr != nil {
					glog.Error(serr)
				}
				return nil, err
			default:
				glog.Error(err)
				_, serr := SetS2iBinaryStatusWithRetry(origin, v1alpha1.StatusUnableToDownload)
				if serr != nil {
					glog.Error(serr)
				}
				return nil, err
			}
		}
		glog.Error(err)
		return nil, err
	}

	if copy.Spec.UploadTimeStamp == nil {
		copy.Spec.UploadTimeStamp = new(metav1.Time)
	}
	*copy.Spec.UploadTimeStamp = metav1.Now()
	resp, err := k8s.KsClient().DevopsV1alpha1().S2iBinaries(namespace).Update(copy)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	resp, err = SetS2iBinaryStatusWithRetry(resp, v1alpha1.StatusReady)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	return resp, nil
}

func DownloadS2iBinary(namespace, name, fileName string) (string, error) {
	origin, err := informers.KsSharedInformerFactory().Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(namespace).Get(name)
	if err != nil {
		glog.Errorf("%+v", err)
		return "", err
	}
	if origin.Spec.FileName != fileName {
		err := fmt.Errorf("could not fould file %s", fileName)
		glog.Error(err)
		return "", err
	}
	if origin.Status.Phase != v1alpha1.StatusReady {
		err := restful.NewError(http.StatusBadRequest, "file is not ready, please try later")
		glog.Error(err)
		return "", err
	}
	s3Client := s2is3.Client()
	if s3Client == nil {
		err := fmt.Errorf("could not get s3 client")
		glog.Error(err)
		return "", err
	}
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     s2is3.Bucket(),
		Key:                        aws.String(fmt.Sprintf("%s-%s", namespace, name)),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", origin.Spec.FileName)),
	})
	url, err := req.Presign(5 * time.Minute)
	if err != nil {
		glog.Error(err)
		return "", err
	}
	return url, nil

}

func SetS2iBinaryStatus(s2ibin *v1alpha1.S2iBinary, status string) (*v1alpha1.S2iBinary, error) {
	copy := s2ibin.DeepCopy()
	copy.Status.Phase = status
	copy, err := k8s.KsClient().DevopsV1alpha1().S2iBinaries(s2ibin.Namespace).Update(copy)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	return copy, nil
}

func SetS2iBinaryStatusWithRetry(s2ibin *v1alpha1.S2iBinary, status string) (*v1alpha1.S2iBinary, error) {

	var bin *v1alpha1.S2iBinary
	var err error
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		bin, err = informers.KsSharedInformerFactory().Devops().V1alpha1().S2iBinaries().Lister().S2iBinaries(s2ibin.Namespace).Get(s2ibin.Name)
		if err != nil {
			glog.Error(err)
			return err
		}
		bin.Status.Phase = status
		bin, err = k8s.KsClient().DevopsV1alpha1().S2iBinaries(s2ibin.Namespace).Update(bin)
		if err != nil {
			glog.Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return bin, nil
}
