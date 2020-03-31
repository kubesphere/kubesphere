package v1alpha3

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "devops.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha3"}

func AddToContainer(c *restful.Container, notifier *v1alpha3.EventNotifier) error {
	eventHandler := NewKubeSphereEventHandler(notifier)
	webservice := runtime.NewWebService(GroupVersion)
	webservice.Route(webservice.POST("/jenkinsEvent/${eventType}").
		To(eventHandler.JenkinsEventHandler).Param(restful.PathParameter(
		"eventType", "event type of event")).
		Reads(v1alpha3.JenkinsEvent{}))
	c.Add(webservice)
	return nil
}

// RegisterEventHandler registers the EventHandler into the notifier.
// When implementing a specific EventHandler, it should be registered from here
func RegisterEventHandler(notifier *v1alpha3.EventNotifier) error {
	//notifier.RegisterEventHandler()
	return nil
}
