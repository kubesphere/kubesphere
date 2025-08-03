// ...existing code...
import (
	// ...existing code...
	"net/http"
	"net/http/httputil"
	"net/url"
	"github.com/emicklei/go-restful/v3"
	// ...existing code...
)

// ...existing code...

func proxyToApiServer(ws *restful.WebService) {
	ws.Path("/proxy")
	ws.Route(ws.GET("/services/{namespace}/{service}/{path:*}").
		To(func(req *restful.Request, resp *restful.Response) {
			proxyURL := fmt.Sprintf("http://ks-apiserver.kubesphere-system.svc/kapis/proxy/v1alpha1/namespaces/%s/services/%s/%s",
				req.PathParameter("namespace"),
				req.PathParameter("service"),
				req.PathParameter("path"))
			target, _ := url.Parse(proxyURL)
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(resp.ResponseWriter, req.Request)
		}).
		Doc("Proxy service through console"))
}

// ...existing code...
// Make sure to call proxyToApiServer(ws) in your route registration

