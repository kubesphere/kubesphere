package v2beta2

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
	"io"
	"net/http"
)

type handler struct {}

type Result struct {
	Code int  `json:"Status"`
	Message string  `json:"Message"`
}

func newHandler() handler{
	return handler{}
}

func (h handler) Verify (request *restful.Request, response *restful.Response) {
	resp, err := http.Post("http://notification-manager-svc.kubesphere-monitoring-system.svc:19093/api/v2/verify", "json", request.Request.Body)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// return 500
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	var result Result
	json.Unmarshal([]byte(body), &result)
	response.WriteAsJson(result)

}

