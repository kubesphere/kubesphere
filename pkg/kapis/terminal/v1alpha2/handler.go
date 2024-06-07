/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	requestctx "kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/terminal"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

type handler struct {
	client          kubernetes.Interface
	config          *rest.Config
	terminaler      terminal.Interface
	authorizer      authorizer.Authorizer
	uploadFileLimit int64
}

func (h *handler) HandleTerminalSession(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	shell := request.QueryParameter("shell")

	user, _ := requestctx.UserFrom(request.Request.Context())

	createPodExec := authorizer.AttributesRecord{
		User:            user,
		Verb:            "create",
		Resource:        "pods",
		Subresource:     "exec",
		Name:            podName,
		Namespace:       namespace,
		ResourceRequest: true,
		ResourceScope:   requestctx.NamespaceScope,
	}

	decision, reason, err := h.authorizer.Authorize(createPodExec)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if decision != authorizer.DecisionAllow {
		api.HandleForbidden(response, request, errors.New(reason))
		return
	}

	conn, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		klog.Warning(err)
		return
	}

	h.terminaler.HandleSession(request.Request.Context(), shell, namespace, podName, containerName, conn)
}

func (h *handler) HandleUserKubectlSession(request *restful.Request, response *restful.Response) {
	user, _ := requestctx.UserFrom(request.Request.Context())

	createPodExec := authorizer.AttributesRecord{
		User:            user,
		Verb:            "create",
		Resource:        "pods",
		Subresource:     "exec",
		Namespace:       constants.KubeSphereNamespace,
		ResourceRequest: true,
		Name:            fmt.Sprintf("%s-%s", constants.KubectlPodNamePrefix, user.GetName()),
		ResourceScope:   requestctx.NamespaceScope,
	}

	decision, reason, err := h.authorizer.Authorize(createPodExec)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if decision != authorizer.DecisionAllow {
		api.HandleForbidden(response, request, errors.New(reason))
		return
	}

	conn, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		klog.Warning(err)
		return
	}
	h.terminaler.HandleUserKubectlSession(request.Request.Context(), user.GetName(), conn)
}

func (h *handler) HandleShellAccessToNode(request *restful.Request, response *restful.Response) {
	nodename := request.PathParameter("nodename")

	user, _ := requestctx.UserFrom(request.Request.Context())

	createNodesExec := authorizer.AttributesRecord{
		User:            user,
		Verb:            "create",
		Resource:        "nodes",
		Subresource:     "exec",
		ResourceRequest: true,
		ResourceScope:   requestctx.ClusterScope,
	}

	decision, reason, err := h.authorizer.Authorize(createNodesExec)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if decision != authorizer.DecisionAllow {
		api.HandleForbidden(response, request, errors.New(reason))
		return
	}

	conn, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		klog.Warning(err)
		return
	}

	h.terminaler.HandleShellAccessToNode(request.Request.Context(), nodename, conn)
}

type fileWithHeader struct {
	file   multipart.File
	header *multipart.FileHeader
}

func (h *handler) UploadFile(request *restful.Request, response *restful.Response) {
	if err := request.Request.ParseMultipartForm(h.uploadFileLimit); err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	files := make([]fileWithHeader, 0)
	for name := range request.Request.MultipartForm.File {
		file, header, err := request.Request.FormFile(name)
		if err != nil {
			api.HandleBadRequest(response, nil, err)
			return
		}
		files = append(files, fileWithHeader{
			file:   file,
			header: header,
		})
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()

		tarWriter := tar.NewWriter(writer)
		defer tarWriter.Close()

		for _, f := range files {
			func(f fileWithHeader) {
				defer f.file.Close()

				// Write the tar header to the tar file
				if err := tarWriter.WriteHeader(&tar.Header{
					Name: f.header.Filename,
					Mode: 0600,
					Size: f.header.Size,
				}); err != nil {
					api.HandleInternalError(response, nil, err)
					return
				}
				// Copy the file content to the tar file
				if _, err := io.Copy(tarWriter, f.file); err != nil {
					api.HandleInternalError(response, nil, err)
					return
				}
			}(f)
		}
	}()

	targetDir := request.QueryParameter("path")
	if targetDir == "" {
		targetDir = "/"
	}

	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")

	req := h.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"tar", "-xmf", "-", "-C", targetDir},
			Stdin:     true,
			Stdout:    false,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(h.config, "POST", req.URL())
	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	if err = exec.StreamWithContext(request.Request.Context(), remotecommand.StreamOptions{
		Stdin:             reader,
		Stdout:            nil,
		Stderr:            response,
		TerminalSizeQueue: nil,
	}); err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}
}

type tarPipe struct {
	config *rest.Config
	client rest.Interface

	reader    *io.PipeReader
	outStream *io.PipeWriter
	bytesRead uint64
	size      uint64
	ctx       context.Context

	namespace, name, container, filePath string
}

func newTarPipe(ctx context.Context, config *rest.Config, client rest.Interface, namespace, name, container, filePath string) (*tarPipe, error) {
	t := &tarPipe{
		config:    config,
		client:    client,
		namespace: namespace,
		name:      name,
		container: container,
		filePath:  filePath,
		ctx:       ctx,
	}

	if err := t.getFileSize(); err != nil {
		return nil, err
	}
	if err := t.initReadFrom(0); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *tarPipe) getFileSize() error {
	req := t.client.Post().
		Resource("pods").
		Name(t.name).
		Namespace(t.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: t.container,
			Command:   []string{"sh", "-c", fmt.Sprintf("tar cf - %s | wc -c", t.filePath)},
			Stdin:     false,
			Stdout:    true,
			Stderr:    false,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()

		if err = exec.StreamWithContext(t.ctx, remotecommand.StreamOptions{
			Stdin:             nil,
			Stdout:            writer,
			Stderr:            nil,
			TerminalSizeQueue: nil,
		}); err != nil {
			klog.Error(err)
		}
	}()

	result, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	num, err := strconv.ParseUint(strings.TrimSpace(string(result)), 10, 64)
	if err != nil {
		return err
	}
	t.size = num
	return nil
}

func (t *tarPipe) initReadFrom(n uint64) error {
	t.reader, t.outStream = io.Pipe()

	req := t.client.Post().
		Resource("pods").
		Name(t.name).
		Namespace(t.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: t.container,
			Command:   []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", t.filePath, n)},
			Stdin:     false,
			Stdout:    true,
			Stderr:    false,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
	if err != nil {
		return err
	}

	go func() {
		defer t.outStream.Close()

		if err = exec.StreamWithContext(t.ctx, remotecommand.StreamOptions{
			Stdin:             nil,
			Stdout:            t.outStream,
			Stderr:            nil,
			TerminalSizeQueue: nil,
		}); err != nil {
			klog.Error(err)
		}
	}()
	return nil
}

func (t *tarPipe) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)
	if err != nil {
		if t.bytesRead == t.size {
			return n, io.EOF
		}
		return n, t.initReadFrom(t.bytesRead + 1)
	}
	t.bytesRead += uint64(n)
	return n, nil
}

func (h *handler) DownloadFile(request *restful.Request, response *restful.Response) {
	filePath := request.QueryParameter("path")
	fileName := filepath.Base(filePath)

	response.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tar", fileName))

	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")

	reader, err := newTarPipe(request.Request.Context(), h.config, h.client.CoreV1().RESTClient(), namespace, podName, containerName, filePath)
	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	if _, err = io.Copy(response.ResponseWriter, reader); err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}
}
