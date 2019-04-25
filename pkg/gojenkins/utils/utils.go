package utils

import (
	"github.com/asaskevich/govalidator"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"net/http"
	"strconv"
)

func GetJenkinsStatusCode(jenkinsErr error) int {
	if code, err := strconv.Atoi(jenkinsErr.Error()); err == nil {
		message := http.StatusText(code)
		if !govalidator.IsNull(message) {
			return code
		}
	}
	if jErr, ok := jenkinsErr.(*gojenkins.ErrorResponse); ok {
		return jErr.Response.StatusCode
	}
	return http.StatusInternalServerError
}
