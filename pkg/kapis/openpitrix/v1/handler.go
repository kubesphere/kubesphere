package v1

import "github.com/emicklei/go-restful"

type openpitrixHandler struct {
}

func newOpenpitrixHandler() *openpitrixHandler {
	return &openpitrixHandler{}
}

func (h *openpitrixHandler) handleListApplications(request *restful.Request, response *restful.Response) {

}
