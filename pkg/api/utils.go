package api

import (
	"github.com/emicklei/go-restful"
	"net/http"
)

func HandleInternalError(response *restful.Response, err error) {
	statusCode := http.StatusInternalServerError

	response.WriteError(statusCode, err)
}

func HandleBadRequest(response *restful.Response, err error) {

}

func HandleNotFound(response *restful.Response, err error) {

}

func HandleForbidden(response *restful.Response, err error) {

}
