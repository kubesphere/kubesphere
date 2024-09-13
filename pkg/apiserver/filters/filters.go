/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/apiserver/pkg/endpoints/responsewriter"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/metrics"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
)

var _ http.ResponseWriter = &metaResponseWriter{}
var _ responsewriter.UserProvidedDecorator = &metaResponseWriter{}

type metaResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *metaResponseWriter) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func newMetaResponseWriter(w http.ResponseWriter) *metaResponseWriter {
	return &metaResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *metaResponseWriter) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *metaResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *metaResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	if err != nil {
		return size, err
	}
	return size, nil
}

func (r *metaResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
	}
	return hijacker.Hijack()
}

func WithGlobalFilter(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		wrapper := newMetaResponseWriter(w)
		start := time.Now()

		handler.ServeHTTP(responsewriter.WrapForHTTP1Or2(wrapper), req)
		elapsedTime := time.Since(start)

		// Record metrics for each request
		reqInfo, exists := request.RequestInfoFrom(req.Context())
		if exists && reqInfo.APIGroup != "" {
			metrics.RequestCounter.WithLabelValues(
				reqInfo.Verb, reqInfo.APIGroup, reqInfo.APIVersion, reqInfo.Resource, strconv.Itoa(wrapper.statusCode),
			).Inc()
			metrics.RequestLatencies.WithLabelValues(
				reqInfo.Verb, reqInfo.APIGroup, reqInfo.APIVersion, reqInfo.Resource,
			).Observe(elapsedTime.Seconds())
		}

		// Record log for each request
		logWithVerbose := klog.V(4)
		// Always log error response
		if wrapper.statusCode > http.StatusBadRequest {
			logWithVerbose = klog.V(0)
		}

		logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
			iputil.RemoteIp(req),
			req.Method,
			req.URL,
			req.Proto,
			wrapper.statusCode,
			wrapper.size,
			elapsedTime.Milliseconds(),
		)
	})
}
