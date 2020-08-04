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

package jenkins

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"net/url"
	"time"
)

// TODO: deprecated, use SendJenkinsRequestWithHeaderResp() instead
func (j *Jenkins) SendPureRequest(path string, httpParameters *devops.HttpParameters) ([]byte, error) {
	resBody, _, err := j.SendPureRequestWithHeaderResp(path, httpParameters)

	return resBody, err
}

// provider request header to call jenkins api.
// transfer bearer token to basic token for inner Oauth and Jeknins
func (j *Jenkins) SendPureRequestWithHeaderResp(path string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	Url, err := url.Parse(j.Server + path)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	header := httpParameters.Header
	SetBasicBearTokenHeader(&header)

	newRequest := &http.Request{
		Method:   httpParameters.Method,
		URL:      Url,
		Header:   header,
		Body:     httpParameters.Body,
		Form:     httpParameters.Form,
		PostForm: httpParameters.PostForm,
	}

	resp, err := client.Do(newRequest)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	resBody, _ := getRespBody(resp)
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		klog.Errorf("%+v", string(resBody))
		jkerr := new(JkError)
		jkerr.Code = resp.StatusCode
		jkerr.Message = string(resBody)
		return nil, nil, jkerr
	}

	return resBody, resp.Header, nil
}
