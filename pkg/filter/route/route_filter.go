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

package route

import (
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

// Route Filter (defines FilterFunction)
func RouteLogging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	start := time.Now()
	chain.ProcessFilter(req, resp)
	glog.Infof("%s - \"%s %s %s\" %d %dms",
		strings.Split(req.Request.RemoteAddr, ":")[0],
		req.Request.Method,
		req.Request.URL.RequestURI(),
		req.Request.Proto,
		resp.StatusCode(),
		time.Now().Sub(start)/1000000,
	)

}
