package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/informers"
	events "kubesphere.io/kubesphere/pkg/models/events"
	evtsclient "kubesphere.io/kubesphere/pkg/simple/client/events"
)

type handler struct {
	eo events.Interface
}

func newHandler(factory informers.InformerFactory, ec evtsclient.Client) *handler {
	return &handler{eo: events.NewEventsOperator(ec)}
}

func (h handler) Query(req *restful.Request, resp *restful.Response) {
}
