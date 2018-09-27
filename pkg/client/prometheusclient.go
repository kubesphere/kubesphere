/*
Copyright 2018 The KubeSphere Authors.
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
package client

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const (
	DefaultScheme            = "http"
	DefaultPrometheusService = "prometheus-k8s.monitoring.svc"
	DefaultPrometheusPort    = "9090"
	PrometheusApiPath        = "/api/v1/"
	PrometheusEndpointUrl    = DefaultScheme + "://" + DefaultPrometheusService + ":" + DefaultPrometheusPort + PrometheusApiPath
)

var client = &http.Client{}

func SendRequest(postfix string, params string) string {
	epurl := PrometheusEndpointUrl + postfix + params
	//glog.Info("monitoring epurl:>", epurl)
	response, err := client.Get(epurl)
	if err != nil {
		glog.Error(err)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			glog.Error(err)
		}
		return string(contents)
	}
	return ""
}

func SendPrometheusRequest(request *restful.Request, recordingRule string) string {
	paramsMap, bol, err := ParseRequestHeader(request)
	if err != nil {
		glog.Error(err)
		return ""
	}

	var res = ""
	var postfix = ""
	if bol {
		// range query
		postfix = "query_range?"
	} else {
		// query
		postfix = "query?"
	}
	paramsMap.Set("query", recordingRule)
	params := paramsMap.Encode()
	res = SendRequest(postfix, params)
	return res
}

func ParseRequestHeader(request *restful.Request) (url.Values, bool, error) {
	instantTime := request.QueryParameter("time")
	start := request.QueryParameter("start")
	end := request.QueryParameter("end")
	step := request.QueryParameter("step")
	timeout := request.QueryParameter("timeout")
	if timeout == "" {
		timeout = "30s"
	}
	// Whether query or query_range request
	u := url.Values{}
	if start != "" && end != "" && step != "" {
		u.Set("start", start)
		u.Set("end", end)
		u.Set("step", step)
		u.Set("timeout", timeout)
		return u, true, nil
	}
	if instantTime != "" {
		u.Set("time", instantTime)
		u.Set("timeout", timeout)
		return u, false, nil
	} else {
		//u.Set("time", strconv.FormatInt(int64(time.Now().Unix()), 10))
		u.Set("timeout", timeout)
		return u, false, nil
	}

	glog.Error("Parse request failed", u)
	return u, false, errors.Errorf("Parse request failed")
}
