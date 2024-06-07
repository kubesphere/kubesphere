/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package filters

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

func WithRequestInfo(next http.Handler, resolver request.RequestInfoResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// KubeSphere supports kube-apiserver proxy requests in multicluster mode. But kube-apiserver
		// stripped all authorization headers. Use custom header to carry token to avoid losing authentication token.
		// We may need a better way. See issue below.
		// https://github.com/kubernetes/kubernetes/issues/38775#issuecomment-277915961
		authorization := req.Header.Get("Authorization")
		if len(authorization) == 0 {
			xAuthorization := req.Header.Get("X-KubeSphere-Authorization")
			if len(xAuthorization) != 0 {
				req.Header.Set("Authorization", xAuthorization)
				req.Header.Del("X-KubeSphere-Authorization")
			}
		}

		// kube-apiserver proxy rejects all proxy requests with dryRun, we had on choice but to
		// replace it with 'dryrun' before proxy and convert it back before send it to kube-apiserver
		// https://github.com/kubernetes/kubernetes/pull/66083
		// See pkg/apiserver/dispatch/dispatch.go for more details
		if len(req.URL.Query()["dryrun"]) != 0 {
			req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "dryrun", "dryRun", 1)
		}

		// kube-apiserver lost query string when proxy websocket requests, there are several issues opened
		// tracking this, like https://github.com/kubernetes/kubernetes/issues/89360. Also there is a promising
		// PR aim to fix this, but it's unlikely it will get merged soon. So here we are again. Put raw query
		// string in Header and extract it on member cluster.
		if rawQuery := req.Header.Get("X-KubeSphere-Rawquery"); len(rawQuery) != 0 && len(req.URL.RawQuery) == 0 {
			req.URL.RawQuery = rawQuery
			req.Header.Del("X-KubeSphere-Rawquery")
		}

		ctx := req.Context()
		info, err := resolver.NewRequestInfo(req)
		if err != nil {
			responsewriters.InternalError(w, req, fmt.Errorf("failed to crate RequestInfo: %v", err))
			return
		}

		*req = *req.WithContext(request.WithRequestInfo(ctx, info))
		next.ServeHTTP(w, req)
	})
}
