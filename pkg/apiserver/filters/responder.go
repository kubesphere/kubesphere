/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"

	"k8s.io/klog/v2"
)

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	reason := fmt.Sprintf("Error while proxying request: %v", err)
	klog.Errorln(reason)
	statusError := errors.StatusError{
		ErrStatus: metav1.Status{
			Code:    http.StatusBadGateway,
			Message: reason,
			Reason:  metav1.StatusReason(http.StatusText(http.StatusBadGateway)),
		},
	}
	responsewriters.WriteRawJSON(http.StatusBadGateway, statusError, w)
}
