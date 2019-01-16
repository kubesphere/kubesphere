package v1alpha2

import (
	"regexp"
	"strconv"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/resources"
)

type res struct {
}

func (res *res) Register(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/{resources}").To(res.Handler))
}

func (res *res) Handler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	resourceName := req.PathParameter("resources")
	conditions := req.QueryParameter("conditions")
	orderBy := req.QueryParameter("orderBy")
	limit := 65535
	offset := 0
	reverse := false
	if b := req.QueryParameter("reverse"); b != "" {
		b, err := strconv.ParseBool(b)
		if err == nil {
			reverse = b
		}
	}

	if groups := regexp.MustCompile(`^limit=(\d+),page=(\d+)$`).FindStringSubmatch(req.QueryParameter("paging")); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		if page < 0 {
			page = 1
		}
		offset = (page - 1) * limit
	}

	result, err := resources.ListResource(namespace, resourceName, conditions, orderBy, reverse, limit, offset)

	if errors.Handler(err, resp) {
		return
	}

	resp.WriteEntity(result)
}
