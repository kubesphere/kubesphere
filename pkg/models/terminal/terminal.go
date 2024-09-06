/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// the code is mainly from:
// 	   https://github.com/kubernetes/dashboard/blob/master/src/app/backend/handler/terminal.go
// thanks to the related developer

package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/kubectl/lease"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// ctrl+d to close terminal.
	endOfTransmission = "\u0004"

	pongWait = 30 * time.Second
	// Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// PtyHandler is what remote command expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// Session implements PtyHandler (using a SockJS connection)
type Session struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

var (
	NodeSessionCounter sync.Map
)

// Message is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type Message struct {
	Op, Data   string
	Rows, Cols uint16
}

// Next handles pty->process resize events
// Called in a loop from remote command as long as the process is running
func (t Session) Next() *remotecommand.TerminalSize {
	size := <-t.sizeChan
	if size.Height == 0 && size.Width == 0 {
		return nil
	}
	return &size
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remote command as long as the process is running
func (t Session) Read(p []byte) (int, error) {
	var msg Message
	if err := t.conn.ReadJSON(&msg); err != nil {
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
// Called from remote command whenever there is any output
func (t Session) Write(p []byte) (int, error) {
	msg, err := json.Marshal(Message{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	if err := t.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return 0, err
	}
	if err = t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// term puts these in the center of the terminal
func (t Session) Toast(p string) error {
	msg, err := json.Marshal(Message{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}
	if err := t.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	if err = t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (t Session) Close(status uint32, reason string) {
	klog.V(4).Infof("terminal session closed: %d %s", status, reason)
	close(t.sizeChan)
	if err := t.conn.Close(); err != nil {
		klog.Warning("failed to close websocket connection: ", err)
	}
}

type Interface interface {
	HandleSession(ctx context.Context, shell, namespace, podName, containerName string, conn *websocket.Conn)
	HandleUserKubectlSession(ctx context.Context, username string, conn *websocket.Conn)
	HandleShellAccessToNode(ctx context.Context, nodename string, conn *websocket.Conn)
}

type terminaler struct {
	client        kubernetes.Interface
	config        *rest.Config
	options       *Options
	leaseOperator *lease.Operator
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
	return &terminaler{client: client, config: config, options: options, leaseOperator: lease.NewOperator(client)}
}

func NewNodeTerminaler(ctx context.Context, nodename string, options *Options, client kubernetes.Interface) (*NodeTerminaler, error) {
	n := &NodeTerminaler{
		Namespace:     constants.KubeSphereNamespace,
		ContainerName: "nsenter",
		Nodename:      nodename,
		PodName:       nodename + "-shell-access",
		Shell:         "sh",
		Privileged:    true,
		Config:        options,
		client:        client,
	}

	node, err := n.client.CoreV1().Nodes().Get(ctx, n.Nodename, metav1.GetOptions{})
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

func (n *NodeTerminaler) getNSEnterPod(ctx context.Context) (*v1.Pod, error) {
	pod, err := n.client.CoreV1().Pods(n.Namespace).Get(ctx, n.PodName, metav1.GetOptions{})
	if err != nil || (pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodPending) {
		// pod has timed out, but has not been cleaned up
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			if err = n.client.CoreV1().Pods(n.Namespace).Delete(ctx, n.PodName, metav1.DeleteOptions{}); err != nil {
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
						Image: n.Config.NodeShellOptions.Image,
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

		if n.Config.NodeShellOptions.Timeout == 0 {
			p.Spec.Containers[0].Args = []string{"tail", "-f", "/dev/null"}
		} else {
			p.Spec.Containers[0].Args = []string{"sleep", strconv.Itoa(n.Config.NodeShellOptions.Timeout)}
		}

		pod, err = n.client.CoreV1().Pods(n.Namespace).Create(ctx, p, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("create pod failed on %s node: %v", n.Nodename, err)
		}
	}

	return pod, nil
}

// startProcess is called by handleAttach
// Executed cmd in the container specified in request and connects it up with the ptyHandler (a session)
func (t *terminaler) startProcess(ctx context.Context, namespace, podName, containerName string, cmd []string, ptyHandler PtyHandler) error {
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

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
}

// isValidShell checks if the shell is allowed
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

func (t *terminaler) getKubectlPod(ctx context.Context, username string) (*corev1.Pod, error) {
	var (
		pod *corev1.Pod
		err error
	)
	podName := fmt.Sprintf("%s-%s", constants.KubectlPodNamePrefix, username)
	// wait for the pod to be ready
	return pod, wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		pod, err = t.client.CoreV1().Pods(constants.KubeSphereNamespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err = t.createKubectlPod(ctx, podName, username)
				if apierrors.IsAlreadyExists(err) {
					return false, nil
				}
				return false, err
			}
			return false, err
		}
		if !pod.DeletionTimestamp.IsZero() {
			return false, nil
		}
		if !isPodReady(pod) {
			return false, nil
		}
		return true, nil
	})
}

func isPodReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (t *terminaler) createKubectlPod(ctx context.Context, podName, username string) (*corev1.Pod, error) {
	if _, err := t.client.CoreV1().Secrets(constants.KubeSphereNamespace).Get(ctx, fmt.Sprintf("kubeconfig-%s", username), metav1.GetOptions{}); err != nil {
		return nil, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "kubectl",
					Image:           t.options.KubectlOptions.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host-time",
							MountPath: "/etc/localtime",
						},
						{
							Name:      "kubeconfig",
							MountPath: "/root/.kube/",
						},
					},
				},
			},
			ServiceAccountName: "kubesphere",
			Volumes: []corev1.Volume{
				{
					Name: "host-time",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/localtime",
						},
					},
				},
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
						SecretName: fmt.Sprintf("kubeconfig-%s", username),
					}},
				},
			},
		},
	}
	return t.client.CoreV1().Pods(constants.KubeSphereNamespace).Create(ctx, pod, metav1.CreateOptions{})
}

func (t *terminaler) HandleSession(ctx context.Context, shell, namespace, podName, containerName string, conn *websocket.Conn) {
	var err error
	validShells := []string{"bash", "sh"}
	session := &Session{conn: conn, sizeChan: make(chan remotecommand.TerminalSize)}

	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		err = t.startProcess(ctx, namespace, podName, containerName, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for _, testShell := range validShells {
			cmd := []string{testShell}
			if err = t.startProcess(ctx, namespace, podName, containerName, cmd, session); err == nil {
				break
			}
		}
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		session.Close(1, err.Error())
		return
	}

	session.Close(0, "Process exited")
}

func (t *terminaler) HandleUserKubectlSession(ctx context.Context, username string, conn *websocket.Conn) {
	pod, err := t.getKubectlPod(ctx, username)
	if err != nil {
		klog.Errorf("get kubectl pod error: %s", err.Error())
		return
	}
	if err = t.leaseOperator.Create(ctx, pod); err != nil {
		klog.Errorf("create lease for pod %s/%s failed: %s", pod.Namespace, pod.Name, err.Error())
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		klog.V(4).Infof("renew lease for user %s", username)
		if err = t.leaseOperator.Renew(ctx, pod.Namespace, pod.Name); err != nil {
			klog.Errorf("renew lease for pod %s/%s failed: %s", pod.Namespace, pod.Name, err)
		}

		klog.V(4).Infof("sending ping packet to %s", conn.RemoteAddr().String())
		if err = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
			klog.V(4).Infof("failed to send ping packet: %s, closing websocket connection", err)
			cancel()
			_ = conn.Close()
		}
	}, pingPeriod)

	conn.SetReadDeadline(time.Now().Add(pongWait)) // nolint
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait)) // nolint
		return nil
	})
	conn.SetCloseHandler(func(code int, text string) error {
		klog.V(4).Infof("websocket connection closed: code %d, %s", code, text)
		if err := conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait)); err != nil {
			klog.Warning("failed to send close message: ", err)
		}
		cancel()
		return nil
	})

	t.HandleSession(ctx, "bash", pod.Namespace, pod.Name, "kubectl", conn)
}

func (t *terminaler) HandleShellAccessToNode(ctx context.Context, nodename string, conn *websocket.Conn) {
	nodeTerminaler, err := NewNodeTerminaler(ctx, nodename, t.options, t.client)
	if err != nil {
		klog.Warning("node terminaler init error: ", err)
		return
	}

	pod, err := nodeTerminaler.getNSEnterPod(ctx)
	if err != nil {
		klog.Warning("get nsenter pod error: ", err)
		return
	}
	if err = nodeTerminaler.WatchPodStatusBeRunning(ctx, pod); err != nil {
		klog.Warning("watching pod status error: ", err)
		return
	}
	if err = t.leaseOperator.Create(ctx, pod); err != nil {
		klog.Errorf("create lease for pod %s/%s failed: %s", pod.Namespace, pod.Name, err)
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		klog.V(4).Infof("renew lease for node %s", nodename)
		if err = t.leaseOperator.Renew(ctx, pod.Namespace, pod.Name); err != nil {
			klog.Errorf("renew lease for pod %s/%s failed: %s", pod.Namespace, pod.Name, err)
		}

		klog.V(4).Infof("sending ping packet to %s", conn.RemoteAddr().String())
		if err = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
			klog.V(4).Infof("failed to send ping packet: %s, closing websocket connection", err)
			cancel()
			_ = conn.Close()
		}
	}, pingPeriod)

	conn.SetReadDeadline(time.Now().Add(pongWait)) // nolint
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait)) // nolint
		return nil
	})
	conn.SetCloseHandler(func(code int, text string) error {
		klog.V(4).Infof("websocket connection closed: code %d, %s", code, text)
		if err := conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait)); err != nil {
			klog.Warning("failed to send close message: ", err)
		}
		cancel()
		return nil
	})

	t.HandleSession(ctx, nodeTerminaler.Shell, nodeTerminaler.Namespace, nodeTerminaler.PodName, nodeTerminaler.ContainerName, conn)
}

func (n *NodeTerminaler) WatchPodStatusBeRunning(ctx context.Context, pod *v1.Pod) error {
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

	return wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, false, func(ctx context.Context) (done bool, err error) {
		pod, err = n.client.CoreV1().Pods(pod.ObjectMeta.Namespace).Get(ctx, pod.ObjectMeta.Name, metav1.GetOptions{})
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
