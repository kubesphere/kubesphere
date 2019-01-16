package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

var None = New(OK, "success")

func (e *Error) Error() string {
	return fmt.Sprintf("error code: %d,message: %s", e.Code, e.Message)
}
func httpStatusCode(e *Error) int {
	switch e.Code {
	case OK:
		return http.StatusOK
	case InvalidArgument:
		return http.StatusBadRequest

	case Internal:
		fallthrough
	case Unknown:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

func Handler(err error, resp *restful.Response) bool {

	if err == nil {
		return false
	}

	glog.Errorln(err, reflect.TypeOf(err))

	resp.WriteHeaderAndEntity(wrapper(err))

	return true
}
func wrapper(err error) (int, interface{}) {
	switch err.(type) {
	case *Error:
	case *json.UnmarshalTypeError:
		err = New(InvalidArgument, err.Error())
	default:
		err = New(Unknown, err.Error())
	}

	return httpStatusCode(err.(*Error)), err
}
