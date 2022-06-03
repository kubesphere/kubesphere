// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package v2beta2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/api/notification/v2beta2"

	nm "kubesphere.io/kubesphere/pkg/simple/client/notification"
)

const (
	VerificationAPIPath = "/api/v2/verify"
)

type handler struct {
	option *nm.Options
}

type Result struct {
	Code    int    `json:"Status"`
	Message string `json:"Message"`
}
type notification struct {
	Config   v2beta2.Config   `json:"config"`
	Receiver v2beta2.Receiver `json:"receiver"`
}

func newHandler(option *nm.Options) *handler {
	return &handler{
		option,
	}
}

func (h handler) Verify(request *restful.Request, response *restful.Response) {
	opt := h.option
	if opt == nil || len(opt.Endpoint) == 0 {
		response.WriteAsJson(Result{
			http.StatusBadRequest,
			"Cannot find Notification Manager endpoint",
		})
	}
	host := opt.Endpoint
	notification := notification{}
	reqBody, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	err = json.Unmarshal(reqBody, &notification)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	receiver := notification.Receiver
	user := request.PathParameter("user")

	if receiver.Labels["type"] == "tenant" {
		if user != receiver.Labels["user"] {
			response.WriteAsJson(Result{
				http.StatusForbidden,
				"Permission denied",
			})
			return
		}
	}
	if receiver.Labels["type"] == "global" {
		if user != "" {
			response.WriteAsJson(Result{
				http.StatusForbidden,
				"Permission denied",
			})
			return
		}
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", host, VerificationAPIPath), bytes.NewReader(reqBody))
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	req.Header = request.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// return 500
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	var result Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	response.WriteAsJson(result)
}
