/*

  Copyright 2021 The KubeSphere Authors.

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

package v2beta2

import (
	"net/http"

	"github.com/emicklei/go-restful"
	openapi "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
)

const (
	GroupName      = "notification.kubesphere.io"
	KeyOpenAPITags = openapi.KeyOpenAPITags
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v2beta2"}

func AddToContainer(
	container *restful.Container,
	informers informers.InformerFactory,
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface,
	options *notification.Options) error {

	ws := runtime.NewWebService(GroupVersion)
	h := newNotificationHandler(informers, k8sClient, ksClient, options)

	ws.Route(ws.POST("/configs/notification/verification").
		Reads("").
		To(h.Verify).
		Returns(http.StatusOK, api.StatusOK, http.Response{}.Body)).
		Doc("Provide validation for notification-manager information")
	ws.Route(ws.POST("/configs/notification/users/{user}/verification").
		To(h.Verify).
		Param(ws.PathParameter("user", "user name")).
		Returns(http.StatusOK, api.StatusOK, http.Response{}.Body)).
		Doc("Provide validation for notification-manager information")

	// apis for global notification config, receiver, and secret
	ws.Route(ws.GET("/{resources}").
		To(h.ListResource).
		Doc("list the notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include notificationmanagers, configs, receivers, secrets, routers, silences, configmaps")).
		Param(ws.QueryParameter(query.ParameterName, "name used for filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterLabelSelector, "label selector used for filtering").Required(false)).
		Param(ws.QueryParameter("type", "config or receiver type, known values include dingtalk, email, feishu, slack, webhook, wechat").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

	ws.Route(ws.GET("/{resources}/{name}").
		To(h.GetResource).
		Doc("get the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include notificationmanagers, configs, receivers, secrets, routers, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Param(ws.QueryParameter("type", "config or receiver type, known values include dingtalk, feishu, email, slack, webhook, wechat").Required(false)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.POST("/{resources}").
		To(h.CreateResource).
		Doc("create a notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include notificationmanagers, configs, receivers, secrets, routers, silences, configmaps")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.PUT("/{resources}/{name}").
		To(h.UpdateResource).
		Doc("update the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include notificationmanagers, configs, receivers, secrets, routers, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.PATCH("/{resources}/{name}").
		To(h.PatchResource).
		Doc("patch the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include notificationmanagers, configs, receivers, secrets, routers, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.DELETE("/{resources}/{name}").
		To(h.DeleteResource).
		Doc("delete the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, routers, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, errors.None))

	// apis for tenant notification config and receiver
	ws.Route(ws.GET("/users/{user}/{resources}").
		To(h.ListResource).
		Doc("list the notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Param(ws.QueryParameter(query.ParameterName, "name used for filtering").Required(false)).
		Param(ws.QueryParameter(query.ParameterLabelSelector, "label selector used for filtering").Required(false)).
		Param(ws.QueryParameter("type", "config or receiver type, known values include dingtalk, email, feishu, slack, webhook, wechat").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

	ws.Route(ws.GET("/users/{user}/{resources}/{name}").
		To(h.GetResource).
		Doc("get the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Param(ws.QueryParameter("type", "config or receiver type, known values include dingtalk, email, feishu, slack, webhook, wechat").Required(false)).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.POST("/users/{user}/{resources}").
		To(h.CreateResource).
		Doc("create the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.PUT("/users/{user}/{resources}/{name}").
		To(h.UpdateResource).
		Doc("update the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.PATCH("/users/{user}/{resources}/{name}").
		To(h.PatchResource).
		Doc("Patch the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.DELETE("/users/{user}/{resources}/{name}").
		To(h.DeleteResource).
		Doc("delete the specified notification resources").
		Metadata(KeyOpenAPITags, []string{constants.NotificationTag}).
		Param(ws.PathParameter("user", "user name")).
		Param(ws.PathParameter("resources", "known values include configs, receivers, secrets, silences, configmaps")).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource")).
		Returns(http.StatusOK, api.StatusOK, errors.None))

	container.Add(ws)
	return nil
}
