/*
Copyright 2020 KubeSphere Authors

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
	"fmt"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
)

type Interface interface {
	CredentialOperator

	BuildGetter

	PipelineOperator

	ProjectPipelineOperator

	ProjectOperator

	RoleOperator
}

func GetDevOpsStatusCode(devopsErr error) int {
	if code, err := strconv.Atoi(devopsErr.Error()); err == nil {
		message := http.StatusText(code)
		if !govalidator.IsNull(message) {
			return code
		}
	}
	if jErr, ok := devopsErr.(*ErrorResponse); ok {
		return jErr.Response.StatusCode
	}
	return http.StatusInternalServerError
}

type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, e.Response.Request.URL.RequestURI())
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}
