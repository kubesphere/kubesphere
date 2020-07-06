package auditing

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
	auditv1alpha1 "kubesphere.io/kubesphere/pkg/apiserver/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/listers/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"net"
	"net/http"
	"time"
)

const (
	DefaultWebhook       = "kube-auditing-webhook"
	DefaultCacheCapacity = 10000
	CacheTimeout         = time.Second
	SendTimeout          = time.Second * 3
	ChannelCapacity      = 10
)

type Auditing interface {
	Enabled() bool
	K8sAuditingEnabled() bool
	LogRequestObject(req *http.Request, info *request.RequestInfo) *auditv1alpha1.Event
	LogResponseObject(e *auditv1alpha1.Event, resp *ResponseCapture, info *request.RequestInfo)
}

type auditing struct {
	lister  v1alpha1.WebhookLister
	cache   chan *auditv1alpha1.EventList
	backend *Backend
}

func NewAuditing(lister v1alpha1.WebhookLister, url string, stopCh <-chan struct{}) Auditing {

	a := &auditing{
		lister: lister,
		cache:  make(chan *auditv1alpha1.EventList, DefaultCacheCapacity),
	}

	a.backend = NewBackend(url, ChannelCapacity, a.cache, SendTimeout, stopCh)
	return a
}

func (a *auditing) getAuditLevel() audit.Level {
	wh, err := a.lister.Get(DefaultWebhook)
	if err != nil {
		klog.V(8).Info(err)
		return audit.LevelNone
	}

	return (audit.Level)(wh.Spec.AuditLevel)
}

func (a *auditing) Enabled() bool {

	level := a.getAuditLevel()
	if level.Less(audit.LevelMetadata) {
		return false
	}
	return true
}

func (a *auditing) K8sAuditingEnabled() bool {
	wh, err := a.lister.Get(DefaultWebhook)
	if err != nil {
		klog.V(8).Info(err)
		return false
	}

	return wh.Spec.K8sAuditingEnabled
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
//
func (a *auditing) LogRequestObject(req *http.Request, info *request.RequestInfo) *auditv1alpha1.Event {

	e := &auditv1alpha1.Event{
		Workspace: info.Workspace,
		Cluster:   info.Cluster,
		Event: audit.Event{
			RequestURI:               info.Path,
			Verb:                     info.Verb,
			Level:                    a.getAuditLevel(),
			AuditID:                  types.UID(uuid.New().String()),
			Stage:                    audit.StageResponseComplete,
			ImpersonatedUser:         nil,
			UserAgent:                req.UserAgent(),
			RequestReceivedTimestamp: v1.NewMicroTime(time.Now()),
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

	// Handle the devops request which request url matched /devops/{devops}/kind.
	if len(info.Parts) >= 3 && info.Parts[0] == "devops" {
		e.ObjectRef.Subresource = ""
		e.Devops = info.Parts[1]
		// set resource as kind
		e.ObjectRef.Resource = info.Parts[2]

		// If the request url matched /devops/{devops}/kind/{kind}, set resource name as {kind}
		if len(info.Parts) >= 4 {
			e.ObjectRef.Name = info.Parts[3]
		}
	}

	ips := make([]string, 1)
	ips[0] = iputil.RemoteIp(req)
	e.SourceIPs = ips

	user, ok := request.UserFrom(req.Context())
	if ok {
		e.User.Username = user.GetName()
		e.User.UID = user.GetUID()
		e.User.Groups = user.GetGroups()

		for k, v := range user.GetExtra() {
			e.User.Extra[k] = v
		}
	}

	if (e.Level.GreaterOrEqual(audit.LevelRequest) || e.Verb == "create") && req.ContentLength > 0 {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			klog.Error(err)
			return e
		}
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		if e.Level.GreaterOrEqual(audit.LevelRequest) {
			e.RequestObject = &runtime.Unknown{Raw: body}
		}

		// For resource creating request, get resource name from the request body.
		if info.Verb == "create" {
			obj := &auditv1alpha1.Object{}
			if err := json.Unmarshal(body, obj); err == nil {
				e.ObjectRef.Name = obj.Name
			}
		}
	}

	return e
}

func (a *auditing) LogResponseObject(e *auditv1alpha1.Event, resp *ResponseCapture, info *request.RequestInfo) {

	e.StageTimestamp = v1.NewMicroTime(time.Now())
	e.ResponseStatus = &v1.Status{Code: int32(resp.StatusCode())}
	if e.Level.GreaterOrEqual(audit.LevelRequestResponse) {
		e.ResponseObject = &runtime.Unknown{Raw: resp.Bytes()}
	}

	a.cacheEvent(*e)
}

func (a *auditing) cacheEvent(e auditv1alpha1.Event) {

	eventList := &auditv1alpha1.EventList{}
	eventList.Items = append(eventList.Items, e)
	select {
	case a.cache <- eventList:
		return
	case <-time.After(CacheTimeout):
		klog.Errorf("cache audit event %s timeout", e.AuditID)
		break
	}
}

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

func (c *ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *ResponseCapture) Write(data []byte) (int, error) {

	c.WriteHeader(http.StatusOK)
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
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
	return c.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
