/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	restapi "kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/terminal"
)

const (
	GroupName = "terminal.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func NewHandler(client kubernetes.Interface, authorizer authorizer.Authorizer, config *rest.Config, options *terminal.Options) restapi.Handler {
	var uploadFileLimit int64 = 100 << 20 // 100 MB
	q, err := resource.ParseQuantity(options.UploadFileLimit)
	if err != nil {
		klog.Warningf("parse UploadFileLimit failed: %s, using default value 100Mi", err.Error())
	} else {
		uploadFileLimit = q.Value()
	}

	return &handler{
		client:          client,
		config:          config,
		authorizer:      authorizer,
		terminaler:      terminal.NewTerminaler(client, config, options),
		uploadFileLimit: uploadFileLimit,
	}
}

func NewFakeHandler() restapi.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/exec").
		To(h.HandleTerminalSession).
		Doc("Create pod terminal session").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}).
		Operation("create-pod-exec").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("pod", "pod name")))

	ws.Route(ws.POST("/namespaces/{namespace}/pods/{pod}/file").
		To(h.UploadFile).
		Doc("Upload files to pod").
		Consumes(runtime.MimeMultipartFormData).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}).
		Operation("upload-file-to-pod").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("pod", "pod name")).
		Param(ws.QueryParameter("container", "container name")).
		Param(ws.QueryParameter("path", "dest dir path")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/file").
		To(h.DownloadFile).
		Doc("Download file from pod").
		Consumes(runtime.MimeMultipartFormData).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}).
		Operation("download-file-from-pod").
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("pod", "pod name")).
		Param(ws.QueryParameter("container", "container name")).
		Param(ws.QueryParameter("path", "file path")).
		Returns(http.StatusOK, api.StatusOK, nil))

	ws.Route(ws.GET("/users/{user}/kubectl").
		To(h.HandleUserKubectlSession).
		Param(ws.PathParameter("user", "username")).
		Doc("Create kubectl pod terminal session for current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}).
		Operation("create-user-kubectl-pod-exec"))

	// Add new Route to support shell access to the node
	ws.Route(ws.GET("/nodes/{nodename}/exec").
		To(h.HandleShellAccessToNode).
		Doc("Create node terminal session").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}).
		Operation("create-node-exec").
		Param(ws.PathParameter("nodename", "node name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagTerminal}))

	c.Add(ws)

	return nil
}
