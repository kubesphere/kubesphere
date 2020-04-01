package v1alpha2

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	restful "github.com/emicklei/go-restful"
)

const ScopeQueryUrl = "http://weave-scope-app.weave.svc/api/topology/services"

func getNamespaceTopology(request *restful.Request, response *restful.Response) {
	var query = url.Values{
		"namespace": []string{request.PathParameter("namespace")},
		"timestamp": request.QueryParameters("timestamp"),
	}
	var u = fmt.Sprintf("%s?%s", ScopeQueryUrl, query.Encode())

	resp, err := http.Get(u)

	if err != nil {
		log.Printf("query scope faile with err %v", err)
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("read response error : %v", err)
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		log.Printf("write response failed %v", err)
	}
}

func getNamespaceNodeTopology(request *restful.Request, response *restful.Response) {
	var query = url.Values{
		"namespace": []string{request.PathParameter("namespace")},
		"timestamp": request.QueryParameters("timestamp"),
	}
	var u = fmt.Sprintf("%s/%s?%s", ScopeQueryUrl, request.PathParameter("node_id"), query.Encode())

	resp, err := http.Get(u)

	if err != nil {
		log.Printf("query scope faile with err %v", err)
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("read response error : %v", err)
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		log.Printf("write response failed %v", err)
	}
}
