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

package errors

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
)

type Error struct {
	Message string `json:"message" description:"error message"`
}

var None = Error{Message: "success"}

func (e Error) Error() string {
	return e.Message
}

func Wrap(err error) error {
	return Error{Message: err.Error()}
}

func New(format string, args ...interface{}) error {
	return Error{Message: fmt.Sprintf(format, args...)}
}

func GetServiceErrorCode(err error) int {
	if svcErr, ok := err.(restful.ServiceError); ok {
		return svcErr.Code
	} else {
		return http.StatusInternalServerError
	}
}
