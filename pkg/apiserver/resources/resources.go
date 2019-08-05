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
package resources

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"net/http"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/params"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NamespacedResourceEvents(req *restful.Request, resp *restful.Response) {
	ws, err := upgrader.Upgrade(resp.ResponseWriter, req.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		glog.Infoln("ws: not a websocket handshake")
		return
	} else if err != nil {
		glog.Infoln("ws: failed to upgrade ", err)
		return
	}
	username := req.HeaderParameter(constants.UserNameHeader)
	namespace := req.PathParameter("namespace")
	session := NewSession(username, ws)
	session.subscribe(namespace)
}

func ListNamespacedResources(req *restful.Request, resp *restful.Response) {
	ListResources(req, resp)
}

func ListResources(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	resourceName := req.PathParameter("resources")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = resources.CreateTime
		reverse = true
	}

	if err != nil {
		glog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := resources.ListResources(namespace, resourceName, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		glog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}
