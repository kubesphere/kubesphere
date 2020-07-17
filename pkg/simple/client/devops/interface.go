package devops

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/http"
	"strconv"
)

type Interface interface {
	CredentialOperator

	BuildGetter

	PipelineOperator

	ProjectPipelineOperator

	ProjectOperator

	RoleOperator
}

func GetDevOpsStatusCode(devopsErr error) int {
	if code, err := strconv.Atoi(devopsErr.Error()); err == nil {
		message := http.StatusText(code)
		if !govalidator.IsNull(message) {
			return code
		}
	}
	if jErr, ok := devopsErr.(*ErrorResponse); ok {
		return jErr.Response.StatusCode
	}
	return http.StatusInternalServerError
}

type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, e.Response.Request.URL.RequestURI())
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}
