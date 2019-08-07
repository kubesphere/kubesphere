package devops

import (
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/errors"
	"net/http"
)

func UploadS2iBinary(req *restful.Request, resp *restful.Response) {
	err := req.Request.ParseMultipartForm(bytefmt.MEGABYTE * 20)
	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if len(req.Request.MultipartForm.File) == 0 {
		err := fmt.Errorf("could not get file from form")
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
	}
	if len(req.Request.MultipartForm.File["binary"]) == 0 {
		err := fmt.Errorf("could not get file from form")
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
	}

}

func DownloadS2iBinary(req *restful.Request, resp *restful.Response) {

}
