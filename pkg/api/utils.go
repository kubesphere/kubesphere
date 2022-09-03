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
	"net/http"
	"runtime"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

// Avoid emitting errors that look like valid HTML. Quotes are okay.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

func HandleInternalError(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, w, req, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusBadRequest, w, req, err)
}

func HandleNotFound(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusNotFound, w, req, err)
}

func HandleForbidden(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusForbidden, w, req, err)
}

func HandleUnauthorized(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, w, req, err)
}

func HandleTooManyRequests(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, w, req, err)
}

func HandleConflict(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusConflict, w, req, err)
}

func HandleServiceUnavailable(w http.ResponseWriter, req *restful.Request, err error) {
	handle(http.StatusServiceUnavailable, w, req, err)
}

func HandleError(w http.ResponseWriter, req *restful.Request, err error) {
	var statusCode int
	switch t := err.(type) {
	case errors.APIStatus:
		statusCode = int(t.Status().Code)
	case restful.ServiceError:
		statusCode = t.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, w, req, err)
}

func handle(statusCode int, w http.ResponseWriter, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	klog.Errorf("%s:%d %v", fn, line, err)
	http.Error(w, sanitizer.Replace(err.Error()), statusCode)
}
