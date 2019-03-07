package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"

	k8sError "k8s.io/apimachinery/pkg/api/errors"

	"github.com/emicklei/go-restful"
	"github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

var None = New(OK, "success")

func (e *Error) Error() string {
	return fmt.Sprintf("error: %v,message: %v", e.Code.String(), e.Message)
}
func (e *Error) HttpStatusCode() int {
	switch e.Code {
	case OK:
		return http.StatusOK
	case InvalidArgument:
		return http.StatusBadRequest
	case AlreadyExists:
		return http.StatusConflict
	case Unavailable:
		return http.StatusServiceUnavailable
	case NotImplement:
		return http.StatusNotImplemented
	case VerifyFailed:
		return http.StatusBadRequest
	case Conflict:
		return http.StatusConflict
	case Internal:
		fallthrough
	case Unknown:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

func New(code Code, message string) error {
	if message == "" {
		message = code.String()
	}
	return &Error{Code: code, Message: message}
}

func HandlerError(err error, resp *restful.Response) bool {

	if err == nil {
		return false
	}

	glog.Errorln(reflect.TypeOf(err), err)

	resp.WriteHeaderAndEntity(wrapper(err))

	return true
}

func wrapper(err error) (int, interface{}) {
	switch err.(type) {
	case *Error:
	case *json.UnmarshalTypeError:
		err = New(InvalidArgument, err.Error())
	case *mysql.MySQLError:
		err = wrapperMysqlError(err.(*mysql.MySQLError))
	case *net.OpError:
		err = New(Internal, err.Error())
	default:
		if k8sError.IsNotFound(err) {
			err = New(NotFound, err.Error())
		} else {
			err = New(Unknown, err.Error())
		}

	}
	return err.(*Error).HttpStatusCode(), err
}

func wrapperMysqlError(sqlError *mysql.MySQLError) error {
	switch sqlError.Number {
	case 1062:
		return New(AlreadyExists, sqlError.Message)
	default:
		return New(Unknown, sqlError.Message)
	}
}

func Wrap(data []byte) error {
	var j map[string]string
	err := json.Unmarshal(data, &j)
	if err != nil {
		return errors.New(string(data))
	} else if message := j["message"]; message != "" {
		return errors.New(message)
	} else if message := j["Error"]; message != "" {
		return errors.New(message)
	} else {
		return errors.New(string(data))
	}
}
