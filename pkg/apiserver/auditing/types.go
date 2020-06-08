package auditing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/client-go/listers/auditregistration/v1alpha1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultAuditSink     = "kube-auditing-webhook-auditsink"
	DefaultCacheCapacity = 10000
	CacheTimeout         = time.Second
	SendTimeout          = time.Second * 3
	ChannelCapacity      = 10
)

type Auditing interface {
	Enable() bool
	LogRequestObject(info *request.RequestInfo, req *http.Request) *Event
	LogResponseObject(e *Event, resp *ResponseCapture)
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
	lister  v1alpha1.AuditSinkLister
	cache   chan *EventList
	backend *Backend
	configs []Config
}

func NewAuditing(lister v1alpha1.AuditSinkLister) Auditing {

	a := &auditing{
		lister:  lister,
		cache:   make(chan *EventList, DefaultCacheCapacity),
		configs: loadAuditingConfig(),
	}

	a.backend = NewBackend(ChannelCapacity, a.cache, SendTimeout, a.lister)
	return a
}

func (a *auditing) getAuditLevel() audit.Level {
	as, err := a.lister.Get(DefaultAuditSink)
	if err != nil {
		return audit.LevelNone
	}

	return (audit.Level)(as.Spec.Policy.Level)
}

func (a *auditing) Enable() bool {

	level := a.getAuditLevel()
	if level.Less(audit.LevelMetadata) {
		return false
	}
	return true
}

func (a *auditing) LogRequestObject(info *request.RequestInfo, req *http.Request) *Event {
	var config *Config
	for _, c := range a.configs {
		if c.regular != nil && c.regular.MatchString(info.Path) && strings.ToLower(c.Method) == strings.ToLower(req.Method) {
			config = &c
			break
		}
	}

	e := &Event{
		Workspace: info.Workspace,
		Cluster:   info.Cluster,
		Event: audit.Event{
			Level:            a.getAuditLevel(),
			AuditID:          types.UID(uuid.New().String()),
			Stage:            audit.StageResponseComplete,
			RequestURI:       info.Path,
			Verb:             info.Verb,
			ImpersonatedUser: nil,
			UserAgent:        req.UserAgent(),
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
			RequestReceivedTimestamp: metav1.NewMicroTime(time.Now()),
			Annotations:              nil,
		},
	}

	body := a.getBody(e.Level, config, req)

	if config != nil {
		if len(config.Verb) > 0 {
			e.Verb = config.Verb
		}
		if len(config.Resource) > 0 {
			e.ObjectRef.Resource = config.Resource
		}
		if len(config.Subresource) > 0 {
			e.ObjectRef.Subresource = config.Subresource
		}
		if len(config.NamPath) > 0 {
			name := a.getName(config, body, req)
			if len(name) > 0 {
				e.ObjectRef.Name = name
			}
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

	if e.Level.GreaterOrEqual(audit.LevelRequest) && req.ContentLength > 0 {
		if body != nil && len(body) > 0 {
			e.RequestObject = &runtime.Unknown{Raw: body}
		}
	}

	return e
}

func (a *auditing) LogResponseObject(e *Event, resp *ResponseCapture) {
	e.StageTimestamp = metav1.NewMicroTime(time.Now())
	e.ResponseStatus = &metav1.Status{Code: int32(resp.StatusCode())}
	if e.Level.GreaterOrEqual(audit.LevelRequestResponse) {
		e.ResponseObject = &runtime.Unknown{Raw: resp.Bytes()}
	}

	a.cacheEvent(*e)
}

func (a *auditing) getBody(level audit.Level, c *Config, req *http.Request) []byte {
	flag := false
	// When audit level is greater than LevelRequest, need to read body.
	if level.GreaterOrEqual(audit.LevelRequest) {
		flag = true
	}

	// When the name will be parsing from body, need to read body.
	if c != nil && len(c.NamPath) > 0 {
		for _, p := range c.NamPath {
			if strings.HasPrefix(p, "body.") {
				flag = true
				break
			}
		}
	}

	if flag {
		bs, err := ioutil.ReadAll(req.Body)
		if err != nil {
			klog.Error(err)
			return nil
		}
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
		return bs
	}

	return nil
}

func (a *auditing) getName(c *Config, body []byte, req *http.Request) string {

	if err := req.ParseForm(); err != nil {
		klog.Error(err)
		return ""
	}

	var fm map[string]interface{}
	name := ""
	for _, p := range c.NamPath {
		if len(p) == 0 {
			return ""
		}

		if strings.HasPrefix(p, "parameter.") {
			param := strings.TrimPrefix(p, "parameter.")
			np := getPathParameter(c.URI, req.URL.Path, param)
			if len(np) == 0 {
				klog.Errorf("%s is nil", param)
				return ""
			}
			name = fmt.Sprintf("%s.%s", name, np)
		} else if strings.HasPrefix(p, "body.") {
			if fm == nil || len(fm) == 0 {
				m := make(map[string]interface{})
				err := json.Unmarshal(body, &m)
				if err != nil {
					klog.Error(err)
					return ""
				}

				fm = Flatten(m)
			}

			key := strings.TrimPrefix(p, "body.")
			np := fm[key]
			if np == nil || len(np.(string)) == 0 {
				klog.Errorf("%s is nil", key)
				return ""
			}
			name = fmt.Sprintf("%s.%s", name, np.(string))
		} else {
			return ""
		}
	}

	name = strings.TrimPrefix(name, ".")
	return name
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

func getPathParameter(uri, url, param string) string {
	value := ""
	uriPaths := strings.Split(uri, "/")
	urlPaths := strings.Split(url, "/")
	for i := range uriPaths {
		if uriPaths[i] == "{"+param+"}" {
			value = urlPaths[i]
		}
	}
	return value
}

func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
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

func (c ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c ResponseCapture) Write(data []byte) (int, error) {

	defer func() {
		c.StopCh <- struct{}{}
	}()

	c.WriteHeader(http.StatusOK)
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
}

func (c ResponseCapture) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.status = statusCode
		c.wroteHeader = true
		//c.ResponseWriter.WriteHeader(statusCode)
	}
}

func (c ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

func (c ResponseCapture) StatusCode() int {
	return c.status
}
