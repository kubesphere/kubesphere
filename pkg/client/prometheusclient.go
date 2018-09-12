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
	"github.com/golang/glog"
	"github.com/emicklei/go-restful"
	"net/url"
	"strconv"
	"time"
	"log"
	"github.com/pkg/errors"
)

const (
	DefaultPrometheusScheme  = "http"
	DefaultPrometheusService = "prometheus-k8s.monitoring.svc.cluster.local" //"heapster"
	DefaultPrometheusPort    = "9090"                                     // use the first exposed port on the service
	PrometheusApiPath        = "/api/v1/"
	PrometheusEndpointUrl    = DefaultPrometheusScheme + "://" + DefaultPrometheusService + ":" + DefaultPrometheusPort + PrometheusApiPath
)

var client = &http.Client{}

func SendRequest(postfix string, params string) string {
	url := PrometheusEndpointUrl + postfix + params
	//fmt.Println("URL:>", url)
	response, err := client.Get(url)
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


func MakeRequestParams(request *restful.Request, recordingRule string) (string, error) {
	paramsMap, bol, err := ParseRequestHeader(request)
	if err != nil {
		return "", err
	}

	var res = ""
	if bol {
		postfix := "query_range?"
		paramsMap.Set("query", recordingRule) // "'" + recordingRule + "'"
		params := paramsMap.Encode()
		res = SendRequest(postfix, params)
	} else {
		// query
		postfix := "query?"
		paramsMap.Set("query", recordingRule)
		params := paramsMap.Encode()
		res = SendRequest(postfix, params)
	}
	return res, nil
}

func ParseRequestHeader(request *restful.Request) (url.Values, bool, error) {
	instantTime := request.HeaderParameter("time")
	start := request.HeaderParameter("start")
	end := request.HeaderParameter("end")
	step := request.HeaderParameter("step")
	timeout := request.HeaderParameter("timeout")
	if timeout == "" {
		timeout = "30s"
	}
	// query请求还是 query_range 请求
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
		u.Set("time", strconv.FormatInt(int64(time.Now().Unix()), 10))
		u.Set("timeout", timeout)
		return u, false, nil
	}

	log.Fatal("Parse request failed");
	return u, false, errors.Errorf("Parse request failed")
}
