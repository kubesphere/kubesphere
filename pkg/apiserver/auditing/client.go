/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auditing

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/modern-go/reflect2"
	v1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/endpoints/responsewriter"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/auditing/internal"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing/log"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing/webhook"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
)

const (
	DefaultCacheCapacity = 10000
	CacheTimeout         = time.Second

	DefaultBatchSize     = 100
	DefaultBatchInterval = time.Second * 3
)

type Auditing interface {
	Enabled() bool
	LogRequestObject(req *http.Request, info *request.RequestInfo) *Event
	LogResponseObject(e *Event, resp *ResponseCapture)
}

type auditing struct {
	k8sClient  k8s.Client
	stopCh     <-chan struct{}
	auditLevel audit.Level
	events     chan *Event
	backend    []internal.Backend

	hostname string
	hostIP   string
	cluster  string

	eventBatchSize     int
	eventBatchInterval time.Duration
}

func NewAuditing(kubernetesClient k8s.Client, opts *Options, stopCh <-chan struct{}) Auditing {

	a := &auditing{
		k8sClient:          kubernetesClient,
		stopCh:             stopCh,
		auditLevel:         opts.AuditLevel,
		events:             make(chan *Event, DefaultCacheCapacity),
		hostname:           os.Getenv("HOSTNAME"),
		hostIP:             getHostIP(),
		eventBatchInterval: opts.EventBatchInterval,
		eventBatchSize:     opts.EventBatchSize,
	}

	if a.eventBatchInterval == 0 {
		a.eventBatchInterval = DefaultBatchInterval
	}

	if a.eventBatchSize == 0 {
		a.eventBatchSize = DefaultBatchSize
	}

	a.cluster = a.getClusterName()

	if opts.WebhookOptions.WebhookUrl != "" {
		a.backend = append(a.backend, webhook.NewBackend(opts.WebhookOptions.WebhookUrl,
			opts.WebhookOptions.EventSendersNum))
	}

	if opts.LogOptions.Path != "" {
		a.backend = append(a.backend, log.NewBackend(opts.LogOptions.Path,
			opts.LogOptions.MaxAge,
			opts.LogOptions.MaxBackups,
			opts.LogOptions.MaxSize))
	}

	go a.Start()

	return a
}

func getHostIP() string {
	addrs, err := net.InterfaceAddrs()
	hostip := ""
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					hostip = ipnet.IP.String()
					break
				}
			}
		}
	}

	return hostip
}

func (a *auditing) getClusterName() string {
	ns, err := a.k8sClient.CoreV1().Namespaces().Get(context.Background(), constants.KubeSphereNamespace, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("get %s error: %s", constants.KubeSphereNamespace, err)
		return ""
	}

	if ns.Annotations != nil {
		return ns.Annotations[clusterv1alpha1.AnnotationClusterName]
	}

	return ""
}

func (a *auditing) getAuditLevel() audit.Level {
	if a.auditLevel != "" {
		return a.auditLevel
	}

	return audit.LevelMetadata
}

func (a *auditing) Enabled() bool {

	level := a.getAuditLevel()
	return !level.Less(audit.LevelMetadata)
}

// If the request is not a standard request, or a resource request,
// or part of the audit information cannot be obtained through url,
// the function that handles the request can obtain Event from
// the context of the request, assign value to audit information,
// including name, verb, resource, subresource, message etc like this.
//
//	info, ok := request.AuditEventFrom(request.Request.Context())
//	if ok {
//		info.Verb = "post"
//		info.Name = created.Name
//	}

func (a *auditing) LogRequestObject(req *http.Request, info *request.RequestInfo) *Event {

	// Ignore the dryRun k8s request.
	if info.IsKubernetesRequest {
		if len(req.URL.Query()["dryRun"]) != 0 {
			klog.V(6).Infof("ignore dryRun request %s", req.URL.Path)
			return nil
		}
	}

	e := &Event{
		HostName:  a.hostname,
		HostIP:    a.hostIP,
		Workspace: info.Workspace,
		Cluster:   a.cluster,
		Event: audit.Event{
			RequestURI:               info.Path,
			Verb:                     info.Verb,
			Level:                    a.getAuditLevel(),
			AuditID:                  types.UID(uuid.New().String()),
			Stage:                    audit.StageResponseComplete,
			ImpersonatedUser:         nil,
			UserAgent:                req.UserAgent(),
			RequestReceivedTimestamp: metav1.NowMicro(),
			Annotations:              nil,
			ObjectRef: &audit.ObjectReference{
				Resource:        info.Resource,
				Namespace:       info.Namespace,
				Name:            info.Name,
				UID:             "",
				APIGroup:        info.APIGroup,
				APIVersion:      info.APIVersion,
				ResourceVersion: info.ResourceScope,
				Subresource:     info.Subresource,
			},
		},
	}

	ips := make([]string, 1)
	ips[0] = iputil.RemoteIp(req)
	e.SourceIPs = ips

	user, ok := request.UserFrom(req.Context())
	if ok {
		e.User.Username = user.GetName()
		e.User.UID = user.GetUID()
		e.User.Groups = user.GetGroups()

		e.User.Extra = make(map[string]v1.ExtraValue)
		for k, v := range user.GetExtra() {
			e.User.Extra[k] = v
		}
	}

	if a.needAnalyzeRequestBody(e, req) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Error(err)
			return e
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		if e.Level.GreaterOrEqual(audit.LevelRequest) {
			e.RequestObject = &runtime.Unknown{Raw: body}
		}

		// For resource creating request, get resource name from the request body.
		if info.Verb == "create" {
			obj := &Object{}
			if err := json.Unmarshal(body, obj); err == nil {
				e.ObjectRef.Name = obj.Name
			}
		}

		// for recording disable and enable user
		if e.ObjectRef.Resource == "users" && e.Verb == "update" {
			u := &v1beta1.User{}
			if err := json.Unmarshal(body, u); err == nil {
				if u.Status.State == v1beta1.UserActive {
					e.Verb = "enable"
				} else if u.Status.State == v1beta1.UserDisabled {
					e.Verb = "disable"
				}
			}
		}
	}

	a.getWorkspace(e)

	return e
}

func (a *auditing) getWorkspace(e *Event) {
	if e.Workspace != "" {
		return
	}

	ns := e.ObjectRef.Namespace
	if e.ObjectRef.Resource == "namespaces" {
		ns = e.ObjectRef.Name
	}
	if ns == "" {
		return
	}

	namespace, err := a.k8sClient.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("get %s error: %s", ns, err)
		return
	}

	if namespace.Labels != nil {
		e.Workspace = namespace.Labels[constants.WorkspaceLabelKey]
	}
}

func (a *auditing) needAnalyzeRequestBody(e *Event, req *http.Request) bool {

	if req.ContentLength <= 0 {
		return false
	}

	if e.Level.GreaterOrEqual(audit.LevelRequest) {
		return true
	}

	if e.Verb == "create" {
		return true
	}

	// for recording disable and enable user
	if e.ObjectRef.Resource == "users" && e.Verb == "update" {
		return true
	}

	return false
}

func (a *auditing) LogResponseObject(e *Event, resp *ResponseCapture) {

	e.StageTimestamp = metav1.NowMicro()
	e.ResponseStatus = &metav1.Status{Code: int32(resp.StatusCode())}
	if e.Level.GreaterOrEqual(audit.LevelRequestResponse) {
		e.ResponseObject = &runtime.Unknown{Raw: resp.Bytes()}
	}

	a.cacheEvent(*e)
}

func (a *auditing) cacheEvent(e Event) {
	select {
	case a.events <- &e:
		return
	case <-time.After(CacheTimeout):
		klog.V(8).Infof("cache audit event %s timeout", e.AuditID)
		break
	}
}

func (a *auditing) Start() {
	for {
		events, exit := a.getEvents()
		if exit {
			break
		}

		if len(events) == 0 {
			continue
		}

		byteEvents := a.eventToBytes(events)
		if len(byteEvents) == 0 {
			continue
		}

		for _, b := range a.backend {
			if reflect2.IsNil(b) {
				continue
			}

			b.ProcessEvents(byteEvents...)
		}
	}
}

func (a *auditing) getEvents() ([]*Event, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), a.eventBatchInterval)
	defer cancel()

	var events []*Event
	for {
		select {
		case event := <-a.events:
			if event == nil {
				break
			}
			events = append(events, event)
			if len(events) >= a.eventBatchSize {
				return events, false
			}
		case <-ctx.Done():
			return events, false
		case <-a.stopCh:
			return nil, true
		}
	}
}

func (a *auditing) eventToBytes(events []*Event) [][]byte {
	var res [][]byte
	for _, event := range events {
		bs, err := json.Marshal(event)
		if err != nil {
			// Normally, the serialization failure is caused by the failure of ResponseObject serialization.
			// To ensure the integrity of the auditing event to the greatest extent,
			// it is necessary to delete ResponseObject and and then try to serialize again.
			if event.ResponseObject != nil {
				event.ResponseObject = nil
				bs, err = json.Marshal(event)
			}
		}

		if err != nil {
			klog.Errorf("serialize audit event error: %s", err)
			continue
		}

		res = append(res, bs)
	}

	return res
}

var _ http.ResponseWriter = &ResponseCapture{}
var _ responsewriter.UserProvidedDecorator = &ResponseCapture{}

type ResponseCapture struct {
	http.ResponseWriter
	wroteHeader bool
	status      int
	body        *bytes.Buffer
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		wroteHeader:    false,
		body:           new(bytes.Buffer),
	}
}

func (c *ResponseCapture) Unwrap() http.ResponseWriter {
	return c.ResponseWriter
}

func (c *ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *ResponseCapture) Write(data []byte) (int, error) {
	c.WriteHeader(http.StatusOK)
	c.body.Write(data)

	n, err := c.ResponseWriter.Write(data)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (c *ResponseCapture) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.status = statusCode
		c.ResponseWriter.WriteHeader(statusCode)
		c.wroteHeader = true
	}
}

func (c *ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

func (c *ResponseCapture) StatusCode() int {
	return c.status
}

// Hijack implements the http.Hijacker interface.  This expands
// the Response to fulfill http.Hijacker if the underlying
// http.ResponseWriter supports it.
func (c *ResponseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := c.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
	}
	return hijacker.Hijack()
}

// CloseNotify is part of http.CloseNotifier interface
func (c *ResponseCapture) CloseNotify() <-chan bool {
	//nolint:staticcheck
	return c.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
