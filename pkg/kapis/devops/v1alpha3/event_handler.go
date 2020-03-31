package v1alpha3

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/devops"
)

type KubeSphereEventHandler struct {
	eventHandler devops.EventHandler
}

func NewKubeSphereEventHandler(notifier *v1alpha3.EventNotifier) *KubeSphereEventHandler {
	return &KubeSphereEventHandler{eventHandler: devops.NewEventHandler(notifier)}
}

func (k *KubeSphereEventHandler) JenkinsEventHandler(request *restful.Request, resp *restful.Response) {

	eventType := request.PathParameter("eventType")
	var event *v1alpha3.JenkinsEvent
	err := request.ReadEntity(&event)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	err = k.eventHandler.HandleJenkinsEvent(eventType, event)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	resp.WriteAsJson(struct {
		Ok bool `json:"ok"`
	}{Ok: true})
	return
}
