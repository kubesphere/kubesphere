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

package api

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"net/http"
	"runtime"
)

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(1)
	klog.Errorf("%s:%d %v", fn, line, err)
	_ = response.WriteError(http.StatusInternalServerError, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(1)
	klog.Errorf("%s:%d %v", fn, line, err)
	_ = response.WriteError(http.StatusBadRequest, err)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(1)
	klog.Errorf("%s:%d %v", fn, line, err)
	_ = response.WriteError(http.StatusNotFound, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(1)
	klog.Errorf("%s:%d %v", fn, line, err)
	_ = response.WriteError(http.StatusForbidden, err)
}

func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(1)
	klog.Errorf("%s:%d %v", fn, line, err)
	_ = response.WriteError(http.StatusConflict, err)
}
