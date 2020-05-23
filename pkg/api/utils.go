package api

import (
	"github.com/emicklei/go-restful"
	"net/http"
)

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	response.WriteError(http.StatusInternalServerError, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	response.WriteError(http.StatusBadRequest, err)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	response.WriteError(http.StatusNotFound, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	response.WriteError(http.StatusForbidden, err)
}

func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	response.WriteError(http.StatusConflict, err)
}
