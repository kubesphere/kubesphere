package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

type terminalHandler struct {
	terminaler terminal.Interface
}

func newTerminalHandler(client kubernetes.Interface, config *rest.Config) *terminalHandler {
	return &terminalHandler{
		terminaler: terminal.NewTerminaler(client, config),
	}
}

func (t *terminalHandler) handleTerminalSession(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	shell := request.QueryParameter("shell")

	conn, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		klog.Warning(err)
		return
	}

	t.terminaler.HandleSession(shell, namespace, podName, containerName, conn)
}
