package auditing

import (
	"bytes"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/listers/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
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
	Enable() bool
	K8sAuditingEnable() bool
	LogRequestObject(req *http.Request) *Event
	LogResponseObject(e *Event, resp *ResponseCapture, info *request.RequestInfo)
}

type Event struct {
	//The workspace which this audit event happened
	Workspace string
	//The devops project which this audit event happened
	Cluster string

	audit.Event
}

type EventList struct {
	Items []Event
}

type auditing struct {
	lister  v1alpha1.WebhookLister
	cache   chan *EventList
	backend *Backend
}

func NewAuditing(lister v1alpha1.WebhookLister) Auditing {

	a := &auditing{
		lister: lister,
		cache:  make(chan *EventList, DefaultCacheCapacity),
	}

	a.backend = NewBackend(ChannelCapacity, a.cache, SendTimeout)
	return a
}

func (a *auditing) getAuditLevel() audit.Level {
	wh, err := a.lister.Get(DefaultWebhook)
	if err != nil {
		klog.Error(err)
		return audit.LevelNone
	}

	return (audit.Level)(wh.Spec.AuditLevel)
}

func (a *auditing) Enable() bool {

	level := a.getAuditLevel()
	if level.Less(audit.LevelMetadata) {
		return false
	}
	return true
}

func (a *auditing) K8sAuditingEnable() bool {
	wh, err := a.lister.Get(DefaultWebhook)
	if err != nil {
		klog.Error(err)
		return false
	}

	return wh.Spec.K8sAuditingEnable
}

func (a *auditing) LogRequestObject(req *http.Request) *Event {
	e := &Event{
		Event: audit.Event{
			Level:                    a.getAuditLevel(),
			AuditID:                  types.UID(uuid.New().String()),
			Stage:                    audit.StageResponseComplete,
			ImpersonatedUser:         nil,
			UserAgent:                req.UserAgent(),
			RequestReceivedTimestamp: v1.NewMicroTime(time.Now()),
			Annotations:              nil,
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

		for k, v := range user.GetExtra() {
			e.User.Extra[k] = v
		}
	}

	if e.Level.GreaterOrEqual(audit.LevelRequest) && req.ContentLength > 0 {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			klog.Error(err)
			return e
		}
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		e.RequestObject = &runtime.Unknown{Raw: body}
	}

	return e
}

func (a *auditing) LogResponseObject(e *Event, resp *ResponseCapture, info *request.RequestInfo) {

	// Auditing should igonre k8s request when k8s auditing is enabled.
	if info.IsKubernetesRequest && a.K8sAuditingEnable() {
		return
	}

	e.StageTimestamp = v1.NewMicroTime(time.Now())
	e.ResponseStatus = &v1.Status{Code: int32(resp.StatusCode())}
	if e.Level.GreaterOrEqual(audit.LevelRequestResponse) {
		e.ResponseObject = &runtime.Unknown{Raw: resp.Bytes()}
	}

	// If the request is not a standard request, or a resource request,
	// or part of the audit information cannot be obtained through url,
	// the function that handles the request can obtain RequestInfo from
	// the context of the request, assign value to audit information,
	// including name, verb, resource, subresource, etc like this.
	//
	//	info, ok := request.RequestInfoFrom(request.Request.Context())
	//	if ok {
	//		info.Verb = "post"
	//		info.Name = created.Name
	//	}
	//
	e.Workspace = info.Workspace
	e.Cluster = info.Cluster
	e.RequestURI = info.Path
	e.Verb = info.Verb
	e.ObjectRef = &audit.ObjectReference{
		Resource:        info.Resource,
		Namespace:       info.Namespace,
		Name:            info.Name,
		UID:             "",
		APIGroup:        info.APIGroup,
		APIVersion:      info.APIVersion,
		ResourceVersion: info.ResourceScope,
		Subresource:     info.Subresource,
	}

	a.cacheEvent(*e)
}

func (a *auditing) cacheEvent(e Event) {
	eventList := &EventList{}
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
	StopCh      chan interface{}
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		wroteHeader:    false,
		body:           new(bytes.Buffer),
		StopCh:         make(chan interface{}, 1),
	}
}

func (c *ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *ResponseCapture) Write(data []byte) (int, error) {

	defer func() {
		c.StopCh <- struct{}{}
	}()

	c.WriteHeader(http.StatusOK)
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
}

func (c *ResponseCapture) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.status = statusCode
		c.wroteHeader = true
	}
}

func (c *ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

func (c *ResponseCapture) StatusCode() int {
	return c.status
}
