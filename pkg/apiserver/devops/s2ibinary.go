/*
Copyright 2019 The KubeSphere Authors.

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
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/utils/hashutil"
	"net/http"
)

func UploadS2iBinary(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("namespace")
	name := req.PathParameter("s2ibinary")

	err := req.Request.ParseMultipartForm(bytefmt.MEGABYTE * 20)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if len(req.Request.MultipartForm.File) == 0 {
		err := restful.NewError(http.StatusBadRequest, "could not get file from form")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if len(req.Request.MultipartForm.File["s2ibinary"]) == 0 {
		err := restful.NewError(http.StatusBadRequest, "could not get file from form")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	if len(req.Request.MultipartForm.File["s2ibinary"]) > 1 {
		err := restful.NewError(http.StatusBadRequest, "s2ibinary should only have one file")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	defer req.Request.MultipartForm.RemoveAll()
	file, err := req.Request.MultipartForm.File["s2ibinary"][0].Open()
	if err != nil {
		klog.Error(err)
		errors.ParseSvcErr(err, resp)
		return
	}
	filemd5, err := hashutil.GetMD5(file)
	if err != nil {
		klog.Error(err)
		errors.ParseSvcErr(err, resp)
		return
	}
	md5, ok := req.Request.MultipartForm.Value["md5"]
	if ok && len(req.Request.MultipartForm.Value["md5"]) > 0 {
		if md5[0] != filemd5 {
			err := restful.NewError(http.StatusBadRequest, fmt.Sprintf("md5 not match, origin: %+v, calculate: %+v", md5[0], filemd5))
			klog.Error(err)
			errors.ParseSvcErr(err, resp)
			return
		}
	}

	s2ibin, err := devops.UploadS2iBinary(ns, name, filemd5, req.Request.MultipartForm.File["s2ibinary"][0])
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(s2ibin)

}

func DownloadS2iBinary(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("namespace")
	name := req.PathParameter("s2ibinary")
	fileName := req.PathParameter("file")
	url, err := devops.DownloadS2iBinary(ns, name, fileName)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	http.Redirect(resp.ResponseWriter, req.Request, url, http.StatusFound)
	return
}
