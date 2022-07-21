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

// the code is mainly from:
// 	   https://github.com/kubernetes/dashboard/blob/master/src/app/backend/handler/terminal.go
// thanks to the related developer

package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// ctrl+d to close terminal.
	endOfTransmission = "\u0004"
)

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// TerminalSession implements PtyHandler (using a SockJS connection)
type TerminalSession struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

var (
	NodeSessionCounter sync.Map
)

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op, Data   string
	Rows, Cols uint16
}

// Next handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		if size.Height == 0 && size.Width == 0 {
			return nil
		}
		return &size
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Read(p []byte) (int, error) {

	var msg TerminalMessage
	err := t.conn.ReadJSON(&msg)
	if err != nil {
		return copy(p, endOfTransmission), err
	}

	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, endOfTransmission), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	t.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err = t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}
	t.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err = t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (t TerminalSession) Close(status uint32, reason string) {
	klog.Warning(status, reason)
	close(t.sizeChan)
	t.conn.Close()
}

type Interface interface {
	HandleSession(shell, namespace, podName, containerName string, conn *websocket.Conn)
	HandleShellAccessToNode(nodename string, conn *websocket.Conn)
}

type terminaler struct {
	client  kubernetes.Interface
	config  *rest.Config
	options *Options
}

type NodeTerminaler struct {
	Nodename      string
	Namespace     string
	PodName       string
	ContainerName string
	Shell         string
	Privileged    bool
	Config        *Options
	client        kubernetes.Interface
}

func NewTerminaler(client kubernetes.Interface, config *rest.Config, options *Options) Interface {
	return &terminaler{client: client, config: config, options: options}
}

func NewNodeTerminaler(nodename string, options *Options, client kubernetes.Interface) (*NodeTerminaler, error) {

	n := &NodeTerminaler{
		Namespace:     "kubesphere-controls-system",
		ContainerName: "nsenter",
		Nodename:      nodename,
		PodName:       nodename + "-shell-access",
		Shell:         "sh",
		Privileged:    true,
		Config:        options,
		client:        client,
	}

	node, err := n.client.CoreV1().Nodes().Get(context.Background(), n.Nodename, metav1.GetOptions{})

	if err != nil {
		return n, fmt.Errorf("getting node error. nodename:%s, err: %v", n.Nodename, err)
	}

	flag := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			flag = true
			break
		}
	}
	if !flag {
		return n, fmt.Errorf("node status error. node: %s", n.Nodename)
	}

	return n, nil
}

func (n *NodeTerminaler) getNSEnterPod() (*v1.Pod, error) {
	pod, err := n.client.CoreV1().Pods(n.Namespace).Get(context.Background(), n.PodName, metav1.GetOptions{})

	if err != nil || (pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodPending) {
		// pod has timed out, but has not been cleaned up
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			err := n.client.CoreV1().Pods(n.Namespace).Delete(context.Background(), n.PodName, metav1.DeleteOptions{})
			if err != nil {
				return pod, err
			}
		}

		var p = &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: n.PodName,
			},
			Spec: v1.PodSpec{
				NodeName:      n.Nodename,
				HostPID:       true,
				HostNetwork:   true,
				RestartPolicy: v1.RestartPolicyNever,
				Containers: []v1.Container{
					{
						Name:  n.ContainerName,
						Image: n.Config.Image,
						Command: []string{
							"nsenter", "-m", "-u", "-i", "-n", "-p", "-t", "1",
						},
						Stdin: true,
						TTY:   true,
						SecurityContext: &v1.SecurityContext{
							Privileged: &n.Privileged,
						},
					},
				},
			},
		}

		if n.Config.Timeout == 0 {
			p.Spec.Containers[0].Args = []string{"tail", "-f", "/dev/null"}
		} else {
			p.Spec.Containers[0].Args = []string{"sleep", strconv.Itoa(n.Config.Timeout)}
		}

		pod, err = n.client.CoreV1().Pods(n.Namespace).Create(context.Background(), p, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("create pod failed on %s node: %v", n.Nodename, err)
		}
	}

	return pod, nil
}

func (n NodeTerminaler) CleanUpNSEnterPod() {
	idx, _ := NodeSessionCounter.Load(n.Nodename)
	atomic.AddInt64(idx.(*int64), -1)

	if *(idx.(*int64)) == 0 {
		err := n.client.CoreV1().Pods(n.Namespace).Delete(context.Background(), n.PodName, metav1.DeleteOptions{})
		if err != nil {
			klog.Warning(err)
		}
	}
}

// startProcess is called by handleAttach
// Executed cmd in the container specified in request and connects it up with the ptyHandler (a session)
func (t *terminaler) startProcess(namespace, podName, containerName string, cmd []string, ptyHandler PtyHandler) error {
	req := t.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

func (t *terminaler) HandleSession(shell, namespace, podName, containerName string, conn *websocket.Conn) {
	var err error
	validShells := []string{"bash", "sh"}

	session := &TerminalSession{conn: conn, sizeChan: make(chan remotecommand.TerminalSize)}

	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		err = t.startProcess(namespace, podName, containerName, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for _, testShell := range validShells {
			cmd := []string{testShell}
			if err = t.startProcess(namespace, podName, containerName, cmd, session); err == nil {
				break
			}
		}
	}

	if err != nil {
		session.Close(2, err.Error())
		return
	}

	session.Close(1, "Process exited")
}

func (t *terminaler) HandleShellAccessToNode(nodename string, conn *websocket.Conn) {

	nodeTerminaler, err := NewNodeTerminaler(nodename, t.options, t.client)
	if err != nil {
		klog.Warning("node terminaler init error: ", err)
		return
	}

	pod, err := nodeTerminaler.getNSEnterPod()
	if err != nil {
		klog.Warning("get nsenter pod error: ", err)
		return
	}

	if err := nodeTerminaler.WatchPodStatusBeRunning(pod); err != nil {
		klog.Warning("watching pod status error: ", err)
		return
	} else {
		t.HandleSession(nodeTerminaler.Shell, nodeTerminaler.Namespace, nodeTerminaler.PodName, nodeTerminaler.ContainerName, conn)
		defer nodeTerminaler.CleanUpNSEnterPod()
	}
}

func (n *NodeTerminaler) WatchPodStatusBeRunning(pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodRunning {
		idx, ok := NodeSessionCounter.Load(n.Nodename)
		if ok {
			atomic.AddInt64(idx.(*int64), 1)
		} else {
			i := int64(1)
			NodeSessionCounter.LoadOrStore(n.Nodename, &i)
		}

		return nil
	}

	return wait.Poll(time.Millisecond*500, time.Second*5, func() (done bool, err error) {
		pod, err = n.client.CoreV1().Pods(pod.ObjectMeta.Namespace).Get(context.Background(), pod.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warning(err)
			return false, err
		}

		if pod.Status.Phase == v1.PodRunning {
			idx, ok := NodeSessionCounter.Load(n.Nodename)
			if ok {
				atomic.AddInt64(idx.(*int64), 1)
			} else {
				i := int64(1)
				NodeSessionCounter.LoadOrStore(n.Nodename, &i)
			}
			return true, nil
		}
		return false, nil
	})
}
