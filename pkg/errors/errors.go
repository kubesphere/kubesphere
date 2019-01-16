/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"

	"github.com/go-ldap/ldap"

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
	case WTF:
		return http.StatusTeapot
	case Unavailable:
		return http.StatusServiceUnavailable
	case Forbidden:
		return http.StatusForbidden
	case NotImplement:
		return http.StatusNotImplemented
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

func HandleError(err error, resp *restful.Response) bool {

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
	case *ldap.Error:
		err = wrapperLdapError(err.(*ldap.Error))
	default:
		if k8sError.IsNotFound(err) {
			err = New(NotFound, err.Error())
		} else {
			err = New(Unknown, err.Error())
		}
	}
	return err.(*Error).HttpStatusCode(), err
}

func wrapperLdapError(err *ldap.Error) error {
	switch err.ResultCode {
	case ldap.LDAPResultNoSuchObject:
		return New(NotFound, ldap.LDAPResultCodeMap[err.ResultCode])
	case ldap.LDAPResultInvalidCredentials:
		return New(Unauthorized, ldap.LDAPResultCodeMap[err.ResultCode])
	default:
		return New(Unknown, ldap.LDAPResultCodeMap[err.ResultCode])
	}
}

func wrapperMysqlError(err *mysql.MySQLError) error {
	switch err.Number {
	case 1062:
		return New(Conflict, err.Message)
	default:
		return New(Unknown, err.Message)
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
